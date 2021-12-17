package kitchen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StockTestSuite struct {
	suite.Suite
}

func TestStockTestSuite(t *testing.T) {
	suite.Run(t, new(StockTestSuite))
}

// -- SUITE

func (suite *StockTestSuite) Test_GIVEN_aValidNameAndPositiveQuantity_WHEN_stockItemIsCreated_THEN_createdSuccessfully() {
	// WHEN
	item, err := NewStockItem("Tomato", 1)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "Tomato", item.Name())
	assert.Equal(suite.T(), uint(1), item.Units())
}

func (suite *StockTestSuite) Test_GIVEN_aBlankNameAndPositiveQuantity_WHEN_stockItemIsCreated_THEN_errorIsReturned() {
	// WHEN
	_, err := NewStockItem("", 1)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidStockItem, err.(Error).Code())
	assert.Equal(suite.T(), "Name must be 1 and 25 characters long", err.(Error).Error())
}

func (suite *StockTestSuite) Test_GIVEN_aValidNameAndZeroQuantity_WHEN_stockItemIsCreated_THEN_errorIsReturned() {
	// WHEN
	_, err := NewStockItem("Cheese", 0)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidStockItem, err.(Error).Code())
	assert.Equal(suite.T(), "Units must be greater than 0", err.(Error).Error())
}
