package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type StockHandlerTestSuite struct {
	suite.Suite
}

func TestStockHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(StockHandlerTestSuite))
}

// -- SETUP

func (suite *StockHandlerTestSuite) SetupTest() {
}

func (suite *StockHandlerTestSuite) TearDownTest() {
	clearTables()
}

// -- SUITE

func (suite *StockHandlerTestSuite) Test_GIVEN_noStock_WHEN_stockIsAdded_THEN_totalStockIsCorrect() {
	NewKafkaSender(testKafkaProducer).
		MustSendAsJSON(app.InventoryDelivery, svc.StockRequest{
			Stock: []svc.StockItemRequest{
				{Name: "Cheese", Units: uint(5)},
				{Name: "Donuts", Units: uint(7)},
			},
		})

	time.Sleep(time.Duration(20) * time.Second)

	// WHEN
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.Router().ServeHTTP(w, r)

	// THEN
	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), `{
		"stock": [{
			"name": "Cheese",
			"units": 5
		}, {
			"name": "Donuts",
			"units": 7
		}]
	}`, w.Body.String())
}
