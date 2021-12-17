package kitchen

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gobuffalo/validate"
)

type ErrorCode uint64

const (
	ErrUnknown ErrorCode = iota + 1000
	ErrDatabaseConnectivity
	ErrDatabaseState

	ErrInvalidStockItem
)

var errorCodeNames = map[ErrorCode]string{
	ErrUnknown:              "UNKOWN",
	ErrDatabaseConnectivity: "DATABASE_CONNECTIVITY",
	ErrDatabaseState:        "DATABASE_STATE",
	ErrInvalidStockItem:     "INVALID_STOCK_ITEM",
}

func (c ErrorCode) Name() string {
	var name string
	var ok bool
	if name, ok = errorCodeNames[c]; !ok {
		log.Fatalf("FATAL: No name for error code %d", c)
	}
	return name
}

func (c ErrorCode) Status() int {
	switch c {
	case ErrInvalidStockItem:
		return http.StatusBadRequest

	case ErrDatabaseConnectivity:
		fallthrough
	case ErrDatabaseState:
		fallthrough
	case ErrUnknown:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

type Error interface {
	Code() ErrorCode
	Cause() error
	Error() string
	Fields() map[string]string
}

type internalError struct {
	code    ErrorCode
	cause   error
	message string
	fields  map[string]string
}

func (e internalError) Code() ErrorCode {
	return e.code
}

func (e internalError) Cause() error {
	return e.cause
}

func (e internalError) Error() string {
	return e.message
}

func (e internalError) Fields() map[string]string {
	return e.fields
}

func NewError(code ErrorCode, message string, cause error) Error {
	return &internalError{
		code:    code,
		cause:   fmt.Errorf("%w", cause),
		message: message,
		fields:  map[string]string{},
	}
}

func NewErrorWithFields(code ErrorCode, message string, cause error, fields map[string]string) Error {
	return &internalError{
		code:    code,
		cause:   fmt.Errorf("%w", cause),
		message: message,
		fields:  fields,
	}
}

func makeCoreValidationError(code ErrorCode, errors *validate.Errors) error {
	if !errors.HasAny() {
		return nil
	}

	flatErrors := map[string]string{}
	for field, violations := range errors.Errors {
		flatErrors[field] = strings.Join(violations, ", ")
	}

	listErrors := []string{}
	for _, violations := range flatErrors {
		listErrors = append(listErrors, violations)
	}
	sort.Strings(listErrors)

	return NewErrorWithFields(code, strings.Join(listErrors, ", "), nil,
		flatErrors)
}
