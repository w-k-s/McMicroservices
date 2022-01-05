package test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type OrderHandlerTestSuite struct {
	suite.Suite
	sender KafkaSender
}

func TestOrderStockHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrderHandlerTestSuite))
}

// -- SETUP

func (suite *OrderHandlerTestSuite) SetupTest() {
	testKafkaConsumer.SubscribeTopics([]string{
		app.CreateOrder,
		string(app.OrderReady),
		string(app.OrderFailed),
	}, nil)
	suite.sender = NewKafkaSender(testKafkaProducer)
}

func (suite *OrderHandlerTestSuite) TearDownTest() {
	var err error
	if err = testKafkaConsumer.Unsubscribe(); err != nil {
		log.Fatalf("Failed to unsubscribe testKafkaConsumer in OrderHandlerTestSuite: %s", err)
	}
	if err = ClearTables(); err != nil {
		log.Fatalf("Failed to clear tables in OrderHandlerTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *OrderHandlerTestSuite) Test_GIVEN_sufficientStock_WHEN_orderIsReceived_THEN_orderIsProcessedSuccessfully() {

	// GIVEN
	receiver := NewKafkaReceiver(testKafkaConsumer)
	receiver.Listen()
	defer receiver.Close()

	suite.sender.MustSendAsJSON(app.InventoryDelivery, svc.StockRequest{
		Stock: []svc.StockItemRequest{
			{Name: "Tomatoes", Units: 2},
			{Name: "Onions", Units: 2},
			{Name: "Mustard", Units: 2},
		},
	})

	time.Sleep(time.Duration(25) * time.Second)

	// WHEN
	orderId := uint64(time.Now().Unix())
	suite.sender.MustSendAsJSON(app.CreateOrder, svc.OrderRequest{
		OrderId: orderId,
		Toppings: []string{
			"Tomatoes",
			"Onions",
			"Mustard",
		},
	})

	time.Sleep(time.Duration(40) * time.Second)

	// THEN -- order prepared
	actual := receiver.FirstMessage(string(app.OrderReady))
	expected := fmt.Sprintf("{\"id\":%d,\"status\":\"READY\"}", orderId)
	assert.JSONEq(suite.T(), expected, actual)

	// THEN -- stock updated
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.Router().ServeHTTP(w, r)
	var stockResponse svc.StockResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &stockResponse))
	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), 3, len(stockResponse.Stock))
	assert.Equal(suite.T(), "Mustard", stockResponse.Stock[0].Name)
	assert.Equal(suite.T(), uint(1), stockResponse.Stock[0].Units)
	assert.Equal(suite.T(), "Onions", stockResponse.Stock[1].Name)
	assert.Equal(suite.T(), uint(1), stockResponse.Stock[1].Units)
	assert.Equal(suite.T(), "Tomatoes", stockResponse.Stock[2].Name)
	assert.Equal(suite.T(), uint(1), stockResponse.Stock[2].Units)
}

func (suite *OrderHandlerTestSuite) Test_GIVEN_insufficientStock_WHEN_orderIsReceived_THEN_orderIsRejected() {

	// GIVEN
	receiver := NewKafkaReceiver(testKafkaConsumer)
	receiver.Listen()
	defer receiver.Close()

	// WHEN
	suite.sender.MustSendAsJSON(app.CreateOrder, svc.OrderRequest{
		OrderId: uint64(time.Now().Unix()),
		Toppings: []string{
			"Tomatoes",
			"Onions",
			"Mustard",
		},
	})

	time.Sleep(time.Duration(25) * time.Second)

	// THEN
	message := receiver.FirstMessage(string(app.OrderFailed))
	assert.NotEmpty(suite.T(), message, "Expected Topic to contain a message but no message was published")
	assert.JSONEq(suite.T(), `{
		"detail":"Insufficient stock of \"Tomatoes\"",
		"status":400,
		"title":"INSUFFICIENT_STOCK",
		"type":"/api/v1/problems/2006"
	}`, message)
}
