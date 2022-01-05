package test

import (
	"encoding/json"
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
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down StockHandlerTestSuite: %s", err)
	}
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
	var stockResponse svc.StockResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &stockResponse))
	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), 2, len(stockResponse.Stock))
	assert.Equal(suite.T(), "Cheese", stockResponse.Stock[0].Name)
	assert.Equal(suite.T(), uint(5), stockResponse.Stock[0].Units)
	assert.Equal(suite.T(), "Donuts", stockResponse.Stock[1].Name)
	assert.Equal(suite.T(), uint(7), stockResponse.Stock[1].Units)
}
