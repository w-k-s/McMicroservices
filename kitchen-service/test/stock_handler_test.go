package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/app"
	"github.com/w-k-s/McMicroservices/kitchen-service/internal/broker"
)

// -- SUITE

func Test_GIVEN_noStock_WHEN_stockIsAdded_THEN_totalStockIsCorrect(t *testing.T) {
	// GIVEN
	testConsumer, _ := broker.MockKafkaConsumer(t, map[string]int32{
		app.TopicCreateOrder:       0,
		app.TopicInventoryDelivery: 0,
	},testConfig.Broker().ConsumerConfig())

	testProducer,_ := broker.MockKafkaProducer(t)
	testProducer.VerifyMessageSent(func(body []byte) error {
		expected := "{\"id\":1,\"status\":\"FAILED\",\"reason\":\"insufficient stock of \\\"Tomatoes\\\"\"}\n"
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
	
	// GIVEN
	testConsumer.
		YieldMessage(broker.Message{
			Topic:     app.TopicInventoryDelivery,
			Content:     []byte(`{"stock":[{"name":"Cheese","units":5},{"name":"Donuts","units":7}]}`),
		})

	time.Sleep(5 * time.Second)

	// THEN
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.ServeHTTP(w, r)

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
