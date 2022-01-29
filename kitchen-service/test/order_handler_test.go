package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	dao "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

const messageWaitTimeout = 2 * time.Minute
const messagePollInterval = 5 * time.Second

type OrderHandlerTestSuite struct {
	suite.Suite
	sender   TestMessageSender
	stockDao dao.StockDao
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
	suite.sender = NewTestMessageSender(testKafkaProducer)
	suite.stockDao = db.MustOpenStockDao(testDB)
}

func (suite *OrderHandlerTestSuite) TearDownTest() {
	var err error
	if err = testKafkaConsumer.Unsubscribe(); err != nil {
		log.Fatalf("Failed to unsubscribe testKafkaConsumer in OrderHandlerTestSuite: %s", err)
	}
	log.Println("Tearing down OrderHandlerTestSuite")
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
	receiver := NewTestMessageReceiver(testKafkaConsumer)
	receiver.Listen()
	defer receiver.Close()

	tx := suite.stockDao.MustBeginTx()
	_ = suite.stockDao.Increase(context.Background(), tx, k.Stock{
		k.Must(k.NewStockItem("Tomatoes", 2)),
		k.Must(k.NewStockItem("Onions", 2)),
		k.Must(k.NewStockItem("Mustard", 2)),
	})
	tx.Commit()

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

	// THEN -- order prepared
	receiver.WaitUntilNextMessageInTopic(messageWaitTimeout, string(app.OrderReady)).
		Plus(time.Second)

	log.Printf("\n%q: Received: %s\n", suite.T().Name(), receiver)
	actual := receiver.FirstMessage(string(app.OrderReady))
	expected := fmt.Sprintf("{\"id\":%d,\"status\":\"READY\"}", orderId)
	assert.JSONEq(suite.T(), expected, actual)

	// THEN -- stock updated
	Await(messageWaitTimeout).
		PollEvery(messagePollInterval).
		Until(func() bool {
			tx := suite.stockDao.MustBeginTx()
			defer tx.Commit()
			stock, err := suite.stockDao.Get(context.Background(), tx)
			if err != nil {
				return false
			}
			return len(stock) == 3 && stock[0].Units() == 1
		}).Start()

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
	receiver := NewTestMessageReceiver(testKafkaConsumer)
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

	// THEN
	receiver.WaitUntilNextMessageInTopic(messageWaitTimeout, string(app.OrderFailed)).
		Plus(time.Second)

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
