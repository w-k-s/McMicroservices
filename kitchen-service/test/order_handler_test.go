package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/assert"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

// -- SUITE

func Test_GIVEN_sufficientStock_WHEN_orderIsReceived_THEN_orderIsProcessedSuccessfully(t *testing.T) {

	var (
		stockDao     = db.MustOpenStockDao(testDB)
		testConsumer = mocks.NewConsumer(t, nil)
		testProducer = mocks.NewSyncProducer(t, nil)
		testApp      *app.App
		err          error
	)

	// GIVEN
	tx,_ := stockDao.BeginTx()
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

	testProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(body []byte) error {
		expected := "{\"id\":1,\"status\":\"READY\"}"
		actual := string(body)
		if expected != actual {
			return fmt.Errorf("Expected %q. Got %q", expected, actual)
		}
		return nil
	})

	mockProducerFactory := func(brokerConfig cfg.BrokerConfig) (sarama.SyncProducer, error) {
		return testProducer, nil
	}

	testConsumer.SetTopicMetadata(map[string][]int32{
		app.TopicCreateOrder:       {0},
		app.TopicInventoryDelivery: {0},
	})
	partitionConsumer := testConsumer.ExpectConsumePartition(app.TopicCreateOrder, 0, sarama.OffsetOldest)
	_ = testConsumer.ExpectConsumePartition(app.TopicInventoryDelivery, 0, sarama.OffsetOldest)
	mockConsumerFactory := func(brokerConfig cfg.BrokerConfig) (sarama.Consumer, error) {
		return testConsumer, nil
	}

	if testApp, err =
		app.NewAppBuilder(testConfig).
			SetConsumerFactory(mockConsumerFactory).
			SetProducerFactory(mockProducerFactory).
			Build(); err != nil {
		log.Fatalf("Failed to initialize application for tests. Reason: %s", err)
	}

	// WHEN
	partitionConsumer.
		YieldMessage(&sarama.ConsumerMessage{
			Topic:     app.TopicCreateOrder,
			Partition: 0,
			Value:     []byte(`{"id":1,"toppings":["Tomatoes","Onions","Mustard"]}`),
		})

	// THEN
	// -- Wait for order to be ready
	time.Sleep(30 * time.Second)

	// -- Check that stock was reduced correctly
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.Router().ServeHTTP(w, r)

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
	var (
		testConsumer = mocks.NewConsumer(t, nil)
		testProducer = mocks.NewSyncProducer(t, nil)
		testApp      *app.App
		err          error
	)

	// GIVEN
	testProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(body []byte) error {
		expected := "{\"id\":1,\"status\":\"FAILED\",\"reason\":\"insufficient stock of \\\"Tomatoes\\\"\"}"
		actual := string(body)
		if expected != actual {
			return fmt.Errorf("Expected %q. Got %q", expected, actual)
		}
		return nil
	})

	mockProducerFactory := func(brokerConfig cfg.BrokerConfig) (sarama.SyncProducer, error) {
		return testProducer, nil
	}

	testConsumer.SetTopicMetadata(map[string][]int32{
		app.TopicCreateOrder:       {0},
		app.TopicInventoryDelivery: {0},
	})
	partitionConsumer := testConsumer.ExpectConsumePartition(app.TopicCreateOrder, 0, sarama.OffsetOldest)
	_ = testConsumer.ExpectConsumePartition(app.TopicInventoryDelivery, 0, sarama.OffsetOldest)
	mockConsumerFactory := func(brokerConfig cfg.BrokerConfig) (sarama.Consumer, error) {
		return testConsumer, nil
	}

	if testApp, err =
		app.NewAppBuilder(testConfig).
			SetConsumerFactory(mockConsumerFactory).
			SetProducerFactory(mockProducerFactory).
			Build(); err != nil {
		log.Fatalf("Failed to initialize application for tests. Reason: %s", err)
	}

	// WHEN
	partitionConsumer.
		YieldMessage(&sarama.ConsumerMessage{
			Topic:     app.TopicCreateOrder,
			Partition: 0,
			Value:     []byte(`{"id":1,"toppings":["Tomatoes","Onions","Mustard"]}`),
		})

	// THEN
	// -- Wait for all expectations to be met
	time.Sleep(30 * time.Second)

	// TearDown
	clearTables()
	testApp.Close()
}
