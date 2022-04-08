package kitchen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobuffalo/validate"
)

type InvalidError struct {
	Cause  error
	Fields map[string]string
}

func (i InvalidError) Unwrap() error {
	return i.Cause
}

func (i InvalidError) Error() string {

	sb := strings.Builder{}
	sb.WriteString(i.Cause.Error())

	if len(i.Fields) > 0 {
		fieldErrors := []string{}
		for _, fieldError := range i.Fields {
			fieldErrors = append(fieldErrors, fieldError)
		}
		sort.Strings(fieldErrors)

		sb.WriteString(". ")
		sb.WriteString(strings.Join(fieldErrors, ", "))
	}

	return sb.String()
}

func (i InvalidError) ErrorTitle() string {
	return "Invalid Request"
}

func (i InvalidError) InvalidFields() map[string]string {
	return i.Fields
}

func invalidErrorWithFields(message string, errors *validate.Errors) error {
	if !errors.HasAny() {
		return nil
	}

	flatErrors := map[string]string{}
	for field, violations := range errors.Errors {
		flatErrors[field] = strings.Join(violations, ", ")
	}

	return InvalidError{Cause: fmt.Errorf(message), Fields: flatErrors}
}

type SystemError struct {
	Cause error
}

func NewSystemError(message string, cause error) error {
	return SystemError{Cause: fmt.Errorf("%s. Reason: %w", message, cause)}
}

func (s SystemError) Unwrap() error {
	return s.Cause
}

func (s SystemError) Error() string {
	return s.Cause.Error()
}

func (s SystemError) IsSystemError() bool {
	return true
}

func (i SystemError) ErrorTitle() string {
	return "System Error"
}
