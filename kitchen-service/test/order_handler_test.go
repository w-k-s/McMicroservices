package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/app"
	"github.com/w-k-s/McMicroservices/kitchen-service/internal/broker"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

// -- SUITE

func Test_GIVEN_sufficientStock_WHEN_orderIsReceived_THEN_orderIsProcessedSuccessfully(t *testing.T) {

	// GIVEN
	stockDao     := db.MustOpenStockDao(testDB)
	tx, _ := stockDao.BeginTx()
	if err = tx.Increase(context.Background(), k.Stock{
		k.Must(k.NewStockItem("Tomatoes", 2)),
		k.Must(k.NewStockItem("Onions", 2)),
		k.Must(k.NewStockItem("Mustard", 2)),
	}); err != nil {
		t.Errorf("Failed to update stock in database. Reason: %q", err)
	}
	if err = tx.Commit(); err != nil {
		t.Errorf("Failed to commit stock update. Reason: %q", err)
	}

	testConsumer, _ := broker.MockKafkaConsumer(t, map[string]int32{
		app.TopicCreateOrder:       0,
		app.TopicInventoryDelivery: 0,
	},testConfig.Broker().ConsumerConfig())

	testProducer,_ := broker.MockKafkaProducer(t)
	testProducer.VerifyMessageSent(func(body []byte) error {
		expected := "{\"id\":1,\"status\":\"READY\"}\n"
		actual := string(body)
		if expected != actual {
			return fmt.Errorf("Expected %q. Got %q", expected, actual)
		}
		return nil
	})

	testApp := app.Must(app.New(app.Builder{
			Config: testConfig,
			Pool: testDB,
			Consumer: testConsumer,
			Producer: testProducer,
		}))

	// WHEN
	testConsumer.
		YieldMessage(broker.Message{
			Topic:     app.TopicCreateOrder,
			Content:     []byte(`{"id":1,"toppings":["Tomatoes","Onions","Mustard"]}`),
		})

	// THEN
	// -- Wait for order to be ready
	time.Sleep(30 * time.Second)

	// -- Check that stock was reduced correctly
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.ServeHTTP(w, r)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{
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

	// TearDown
	clearTables()
	testApp.Close()
}

func Test_GIVEN_insufficientStock_WHEN_orderIsReceived_THEN_orderIsRejected(t *testing.T) {
	// GIVEN
	testConsumer, _ := broker.MockKafkaConsumer(t, map[string]int32{
		app.TopicCreateOrder:       0,
		app.TopicInventoryDelivery: 0,
	},testConfig.Broker().ConsumerConfig())

	testProducer,_ := broker.MockKafkaProducer(t)
	testProducer.VerifyMessageSent(func(body []byte) error {
		expected := "{\"id\":1,\"status\":\"FAILED\",\"reason\":\"insufficient stock of \\\"Tomatoes\\\"\"}"
		actual := string(body)
		if expected != actual {
			return fmt.Errorf("Expected %q. Got %q", expected, actual)
		}
		return nil
	})

	testApp := app.Must(app.New(app.Builder{
		Config: testConfig,
		Pool: testDB,
		Consumer: testConsumer,
		Producer: testProducer,
	}))

	// WHEN
	testConsumer.
		YieldMessage(broker.Message{
			Topic:     app.TopicCreateOrder,
			Content:     []byte(`{"id":1,"toppings":["Tomatoes","Onions","Mustard"]}`),
		})

	// THEN
	// -- Wait for all expectations to be met
	time.Sleep(30 * time.Second)

	// TearDown
	clearTables()
	testApp.Close()
}
