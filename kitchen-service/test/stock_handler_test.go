package test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	dao "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type StockHandlerTestSuite struct {
	suite.Suite
	stockDao dao.StockDao
}

func TestStockHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(StockHandlerTestSuite))
}

// -- SETUP

func (suite *StockHandlerTestSuite) SetupTest() {
	suite.stockDao = testStockDao
}

func (suite *StockHandlerTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down AccountHandlerTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *StockHandlerTestSuite) Test_GIVEN_noStock_WHEN_stockIsAdded_THEN_totalStockIsCorrect() {
	// GIVEN
	ctx := context.Background()
	increaseTx := suite.stockDao.MustBeginTx()

	item1, _ := k.NewStockItem("Cheese", 5)
	item2, _ := k.NewStockItem("Donuts", 7)
	assert.Nil(suite.T(), suite.stockDao.Increase(ctx, increaseTx, k.Stock{item1, item2}), "Increase returned error")
	assert.Nil(suite.T(), increaseTx.Commit(), "Commit returned error")

	// WHEN
	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.Router().ServeHTTP(w, r)

	// THEN
	var stockResponse svc.StockResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &stockResponse))
	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), "Cheese", stockResponse.Stock[0].Name)
	assert.Equal(suite.T(), uint(5), stockResponse.Stock[0].Units)
	assert.Equal(suite.T(), "Donuts", stockResponse.Stock[1].Name)
	assert.Equal(suite.T(), uint(7), stockResponse.Stock[1].Units)
}
