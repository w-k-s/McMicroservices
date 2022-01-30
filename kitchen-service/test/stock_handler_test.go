package test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/assert"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
)

// -- SUITE

func Test_GIVEN_noStock_WHEN_stockIsAdded_THEN_totalStockIsCorrect(t *testing.T) {
	var (
		testConsumer = mocks.NewConsumer(t, nil)
		testProducer = mocks.NewSyncProducer(t, nil)
		testApp      *app.App
		err          error
	)

	// GIVEN
	mockProducerFactory := func(brokerConfig cfg.BrokerConfig) (sarama.SyncProducer, error) {
		return testProducer, nil
	}

	testConsumer.SetTopicMetadata(map[string][]int32{
		app.TopicCreateOrder:       {0},
		app.TopicInventoryDelivery: {0},
	})
	_ = testConsumer.ExpectConsumePartition(app.TopicCreateOrder, 0, sarama.OffsetOldest)
	partitionConsumer := testConsumer.ExpectConsumePartition(app.TopicInventoryDelivery, 0, sarama.OffsetOldest)
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
			Topic:     app.TopicInventoryDelivery,
			Partition: 0,
			Value:     []byte(`{"stock":[{"name":"Cheese","units":5},{"name":"Donuts","units":7}]}`),
		})

	time.Sleep(5 * time.Second)

	// THEN
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.Router().ServeHTTP(w, r)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{
		"stock": [{
			"name": "Cheese",
			"units": 5
		}, {
			"name": "Donuts",
			"units": 7
		}]
	}`, w.Body.String())

	// TearDown
	clearTables()
	testApp.Close()
}
