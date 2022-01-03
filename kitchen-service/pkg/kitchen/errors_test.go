package kitchen

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ErrorTestSuite struct {
	suite.Suite
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

// -- SUITE

func (suite *ErrorTestSuite) Test_GIVEN_errorCode_WHEN_mappedToNumber_THEN_mappingIsCorrect() {
	assert.Equal(suite.T(), uint64(1000), uint64(ErrUnknown))
	assert.Equal(suite.T(), uint64(1001), uint64(ErrDatabaseConnectivity))
	assert.Equal(suite.T(), uint64(1002), uint64(ErrDatabaseState))
	assert.Equal(suite.T(), uint64(1003), uint64(ErrUnmarshalling))
	assert.Equal(suite.T(), uint64(1004), uint64(ErrMarshalling))
	assert.Equal(suite.T(), uint64(1005), uint64(ErrInvalidStockItem))
	assert.Equal(suite.T(), uint64(1006), uint64(ErrInsufficientStock))
}

func (suite *ErrorTestSuite) Test_GIVEN_errorCode_WHEN_mappedToHttpStatus_THEN_mappingIsCorrect() {
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrUnknown.Status())
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrDatabaseConnectivity.Status())
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrDatabaseState.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrInvalidStockItem.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrInsufficientStock.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrUnmarshalling.Status())
}
