package test

import (
	"context"
	"sort"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	dao "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
)

type StockDaoTestSuite struct {
	suite.Suite
	stockDao dao.StockDao
}

func TestStockDaoTestSuite(t *testing.T) {
	suite.Run(t, new(StockDaoTestSuite))
}

// -- SETUP

func (suite *StockDaoTestSuite) SetupTest() {
	suite.stockDao = db.MustOpenStockDao(testDB)
}

// -- TEARDOWN

func (suite *StockDaoTestSuite) TearDownTest() {
	clearTables()
}

// -- SUITE

func (suite *StockDaoTestSuite) Test_GIVEN_noStock_WHEN_stockIsAdded_THEN_totalStockIsCorrect() {
	// GIVEN
	ctx := context.Background()
	increaseTx := suite.stockDao.MustBeginTx()

	// WHEN
	item1, _ := k.NewStockItem("Cheese", 5)
	item2, _ := k.NewStockItem("Donuts", 7)
	assert.Nil(suite.T(), suite.stockDao.Increase(ctx, increaseTx, k.Stock{item1, item2}), "Increase returned error")
	assert.Nil(suite.T(), increaseTx.Commit(), "Commit returned error")

	// THEN
	getTx := suite.stockDao.MustBeginTx()
	stock, err := suite.stockDao.Get(ctx, getTx)
	assert.Nil(suite.T(), getTx.Commit())
	sort.Sort(stock)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Cheese", stock[0].Name())
	assert.Equal(suite.T(), uint(5), stock[0].Units())
	assert.Equal(suite.T(), "Donuts", stock[1].Name())
	assert.Equal(suite.T(), uint(7), stock[1].Units())
}

func (suite *StockDaoTestSuite) Test_GIVEN_stock_WHEN_stockIsAdded_THEN_totalStockIsCorrect() {
	// GIVEN
	ctx := context.Background()
	givenTx := suite.stockDao.MustBeginTx()

	item1, _ := k.NewStockItem("Cheese", 5)
	item2, _ := k.NewStockItem("Donuts", 7)
	assert.Nil(suite.T(), suite.stockDao.Increase(ctx, givenTx, k.Stock{item1, item2}), "Increase returned error")
	assert.Nil(suite.T(), givenTx.Commit(), "Commit returned error")

	// WHEN
	increaseTx := suite.stockDao.MustBeginTx()
	item1Addition, _ := k.NewStockItem("Cheese", 5)
	item2Addition, _ := k.NewStockItem("Donuts", 3)
	assert.Nil(suite.T(), suite.stockDao.Increase(ctx, increaseTx, k.Stock{item1Addition, item2Addition}), "Increase returned error")
	assert.Nil(suite.T(), increaseTx.Commit(), "Commit returned error")

	// THEN
	getTx := suite.stockDao.MustBeginTx()
	stock, err := suite.stockDao.Get(ctx, getTx)
	assert.Nil(suite.T(), getTx.Commit())
	sort.Sort(stock)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Cheese", stock[0].Name())
	assert.Equal(suite.T(), uint(10), stock[0].Units())
	assert.Equal(suite.T(), "Donuts", stock[1].Name())
	assert.Equal(suite.T(), uint(10), stock[1].Units())
}

func (suite *StockDaoTestSuite) Test_GIVEN_stock_WHEN_stockIsDecreased_THEN_totalStockIsCorrect() {
	// GIVEN
	ctx := context.Background()
	givenTx := suite.stockDao.MustBeginTx()

	item1, _ := k.NewStockItem("Cheese", 5)
	item2, _ := k.NewStockItem("Donuts", 7)
	assert.Nil(suite.T(), suite.stockDao.Increase(ctx, givenTx, k.Stock{item1, item2}), "Increase returned error")
	assert.Nil(suite.T(), givenTx.Commit(), "Commit returned error")

	// WHEN
	decreaseTx := suite.stockDao.MustBeginTx()
	item1Decrease, _ := k.NewStockItem("Cheese", 4)
	item2Decrease, _ := k.NewStockItem("Donuts", 2)
	assert.Nil(suite.T(), suite.stockDao.Decrease(ctx, decreaseTx, k.Stock{item1Decrease, item2Decrease}), "Decrease returned error")
	assert.Nil(suite.T(), decreaseTx.Commit(), "Commit returned error")

	// THEN
	getTx := suite.stockDao.MustBeginTx()
	stock, err := suite.stockDao.Get(ctx, getTx)
	assert.Nil(suite.T(), getTx.Commit())
	sort.Sort(stock)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Cheese", stock[0].Name())
	assert.Equal(suite.T(), uint(1), stock[0].Units())
	assert.Equal(suite.T(), "Donuts", stock[1].Name())
	assert.Equal(suite.T(), uint(5), stock[1].Units())
}

func (suite *StockDaoTestSuite) Test_GIVEN_stock_WHEN_stockIsDecreasedBeyondAvailability_THEN_errorIsReturned() {
	// GIVEN
	ctx := context.Background()
	givenTx := suite.stockDao.MustBeginTx()

	item1, _ := k.NewStockItem("Cheese", 5)
	item2, _ := k.NewStockItem("Donuts", 7)
	assert.Nil(suite.T(), suite.stockDao.Increase(ctx, givenTx, k.Stock{item1, item2}), "Increase returned error")
	assert.Nil(suite.T(), givenTx.Commit(), "Commit returned error")

	// WHEN
	decreaseTx := suite.stockDao.MustBeginTx()
	item1Decrease, _ := k.NewStockItem("Cheese", 7)
	item2Decrease, _ := k.NewStockItem("Donuts", 10)
	item3Decrease, _ := k.NewStockItem("Fig", 1)
	err := suite.stockDao.Decrease(ctx, decreaseTx, k.Stock{item1Decrease, item2Decrease, item3Decrease})

	// THEN
	assert.NotNil(suite.T(), err)

	assert.Equal(suite.T(), k.ErrInsufficientStock, err.Code())
	assert.Equal(suite.T(), "Insufficient stock of \"Cheese\"", err.Error())
}
