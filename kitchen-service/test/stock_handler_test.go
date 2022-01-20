package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	dao "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type StockHandlerTestSuite struct {
	suite.Suite
	sender   TestMessageSender
	stockDao dao.StockDao
}

func TestStockHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(StockHandlerTestSuite))
}

// -- SETUP

func (suite *StockHandlerTestSuite) SetupTest() {
	suite.sender = NewTestMessageSender(testAmqpProducer)
	suite.stockDao = db.MustOpenStockDao(testDB)
	clearTables()
}

func (suite *StockHandlerTestSuite) TearDownTest() {

}

// -- SUITE

func (suite *StockHandlerTestSuite) Test_GIVEN_noStock_WHEN_stockIsAdded_THEN_totalStockIsCorrect() {
	// WHEN
	suite.sender.MustSendAsJSON("inventory_delivery", "", svc.StockRequest{
		Stock: []svc.StockItemRequest{
			{Name: "Cheese", Units: uint(5)},
			{Name: "Donuts", Units: uint(7)},
		},
	})

	// THEN
	Await(messageWaitTimeout).
		PollEvery(messagePollInterval).
		Until(func() bool {
			tx := suite.stockDao.MustBeginTx()
			defer tx.Commit()
			stock, err := suite.stockDao.Get(context.Background(), tx)
			if err != nil {
				return false
			}
			return len(stock) == 2
		}).Start()

	r, _ := http.NewRequest("GET", "/kitchen/api/v1/stock", nil)
	w := httptest.NewRecorder()
	testApp.Router().ServeHTTP(w, r)

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
