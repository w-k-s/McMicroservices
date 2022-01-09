package test

import (
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

const messageWaitTimeout = time.Duration(5) * time.Minute

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
		app.InventoryDelivery,
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
}

// -- SUITE

// This method ensures tests are run sequentially (i.e. not in parallel). Running tests in parallel causes flakiness.
func (suite *OrderHandlerTestSuite) Test_orderProcessingSequentially() {

	suite.T().Run("", func(t *testing.T) {
		suite.GIVEN_sufficientStock_WHEN_orderIsReceived_THEN_orderIsProcessedSuccessfully()
	})
	clearTables()

	suite.T().Run("", func(t *testing.T) { suite.GIVEN_insufficientStock_WHEN_orderIsReceived_THEN_orderIsRejected() })
}

func (suite *OrderHandlerTestSuite) GIVEN_sufficientStock_WHEN_orderIsReceived_THEN_orderIsProcessedSuccessfully() {
	log.Printf("%q -------------------\n", suite.T().Name())

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

	receiver.WaitUntilNextMessageInTopic(messageWaitTimeout, app.InventoryDelivery).
		Plus(time.Second)

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

	receiver.WaitUntilNextMessageInTopic(messageWaitTimeout, string(app.OrderReady)).
		Plus(time.Second)

	// THEN -- order prepared
	log.Printf("\n%q: Received: %s\n", suite.T().Name(), receiver)
	actual := receiver.FirstMessage(string(app.OrderReady))
	expected := fmt.Sprintf("{\"id\":%d,\"status\":\"READY\"}", orderId)
	assert.JSONEq(suite.T(), expected, actual)

	// THEN -- stock updated
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.Router().ServeHTTP(w, r)

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), `{
		"stock": [{
			"name": "Mustard",
			"units": 1
		}, {
			"name": "Onions",
			"units": 1
		}, {
			"name": "Tomatoes",
			"units": 1
		}]
	}`, w.Body.String())
	log.Println("-------------------")
}

func (suite *OrderHandlerTestSuite) GIVEN_insufficientStock_WHEN_orderIsReceived_THEN_orderIsRejected() {
	log.Printf("%q -------------------\n", suite.T().Name())

	// GIVEN
	receiver := NewKafkaReceiver(testKafkaConsumer)
	receiver.Listen()
	defer receiver.Close()

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

	receiver.WaitUntilNextMessageInTopic(messageWaitTimeout, string(app.OrderFailed)).
		Plus(time.Second)

	// THEN
	log.Printf("\n%q: Received: %s\n", suite.T().Name(), receiver)
	message := receiver.FirstMessage(string(app.OrderFailed))
	assert.NotEmpty(suite.T(), message, "Expected Topic to contain a message but no message was published")
	assert.JSONEq(suite.T(), fmt.Sprintf(`{
		"reason":"Insufficient stock of \"Tomatoes\"",
		"id":%d,
		"status":"FAILED"
	}`, orderId), message)
	log.Println("-------------------")
}
