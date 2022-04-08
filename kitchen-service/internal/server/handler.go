package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	"schneider.vip/problem"
)

type Handler struct {
}

func (h Handler) MustEncodeJson(w http.ResponseWriter, v interface{}, status int) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	w.WriteHeader(status)
	if err := encoder.Encode(v); err != nil {
		log.Fatalf("Failed to encode json '%v'. Reason: %s", v, err)
	}
}

func (h Handler) DecodeJsonOrSendBadRequest(w http.ResponseWriter, req *http.Request, v interface{}) bool {
	decoder := json.NewDecoder(req.Body)
	decoder.UseNumber()
	if err := decoder.Decode(v); err != nil {
		h.MustEncodeProblem(w, req, k.InvalidError{Cause: fmt.Errorf("failed to parse request. Reason: %w", err)})
		return false
	}
	return true
}

func (h Handler) MustEncodeProblem(w http.ResponseWriter, req *http.Request, err error) {

	log.ErrCtx(req.Context(), err).Msg("Encoding Problem")

	title := errorTitle(err)
	code := url.QueryEscape(title)
	detail := err.Error()
	opts := []problem.Option{}
	status := httpStatus(err)

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

func (h Handler) MustMarshal(result []byte, err error) []byte {
	if err != nil {
		log.Fatalf("Marshal to json failed. Reason: %s", err)
	}
	return result
}
