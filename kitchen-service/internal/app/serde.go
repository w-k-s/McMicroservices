package app

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	"schneider.vip/problem"
)

// Represents a serializer-deserializer that:
// - Deserializes the request based on the Content-Type header
// - Serializes the response based on the Consumes/Accept header
type Serde interface {
	MustSerialize(w http.ResponseWriter, req *http.Request, status int, v interface{})
	Deserialize(w http.ResponseWriter, req *http.Request, v interface{}) bool

	MustSerailizeError(w http.ResponseWriter, req *http.Request, status int, err error)
}

// A dumb serde
// - Assumes the request is encoded in JSON and deserializes accordingly. Does not consider the Content-Type header.
// - Serializes the response as JSON. Does not consider the Accept header.
type JsonSerde struct {
	encoder JSONEncoder
	decoder JSONDecoder
}

func (ser JsonSerde) MustSerialize(w http.ResponseWriter, req *http.Request, status int, v interface{}) {
	w.WriteHeader(status)
	ser.encoder.MustEncode(w, v)
}

func (ser JsonSerde) MustDeserialize(w http.ResponseWriter, req *http.Request, v interface{}) bool {
	if err := ser.decoder.Decode(req.Body, v); err != nil {
		ser.MustSerailizeError(w, req, http.StatusBadRequest, k.InvalidError{Cause: fmt.Errorf("failed to parse request. Reason: %w", err)})
		return false
	}
	return true
}

func (ser JsonSerde) MustSerailizeError(w http.ResponseWriter, req *http.Request, status int, err error) {

	log.ErrCtx(req.Context(), err).Msg("Encoding Problem")

	title := errorTitle(err)
	code := url.QueryEscape(title)
	detail := err.Error()
	opts := []problem.Option{}
	status = httpStatus(err)

	for key, value := range errorFields(err) {
		opts = append(opts, problem.Custom(key, value))
	}

	if _, problemError := problem.New(
		problem.Type(fmt.Sprintf("/api/v1/problems/%s", code)),
		problem.Status(status),
		problem.Instance(req.URL.Path),
		problem.Title(title),
		problem.Detail(detail),
	).
		Append(opts...).
		WriteTo(w); problemError != nil {
		log.ErrCtx(req.Context(), problemError).Msgf("Failed to encode problem '%v'", err)
	}
}

func httpStatus(err error) int {
	if isInvalid(err) {
		return 400
	} else {
		return 500
	}
}

func errorTitle(err error) string {
	type hasErrorTitle interface {
		ErrorTitle() string
	}
	if titledError, ok := err.(hasErrorTitle); ok {
		return titledError.ErrorTitle()
	}
	return ""
}

func isInvalid(err error) bool {
	type hasInvalidFields interface {
		InvalidFields() map[string]string
	}
	if _, ok := err.(hasInvalidFields); ok {
		return true
	}
	return false
}

func errorFields(err error) map[string]string {
	type hasInvalidFields interface {
		InvalidFields() map[string]string
	}
	if fieldsError, ok := err.(hasInvalidFields); ok {
		return fieldsError.InvalidFields()
	}
	return map[string]string{}
}

var DefaultJsonSerde = JsonSerde{JSONEncoder{}, JSONDecoder{}}
