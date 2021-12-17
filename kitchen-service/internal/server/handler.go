package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
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
		h.MustEncodeProblem(w, req, ledger.NewError(ledger.ErrRequestUnmarshallingFailed, "Failed to parse request", err))
		return false
	}
	return true
}

func (h Handler) MustEncodeProblem(w http.ResponseWriter, req *http.Request, err error) {

	log.Printf("Error: %s", err.Error())

	title := ledger.ErrUnknown.Name()
	code := ledger.ErrUnknown
	detail := err.Error()
	opts := []problem.Option{}
	status := ledger.ErrUnknown.Status()

	if coreError, ok := err.(ledger.Error); ok {
		title = coreError.Code().Name()
		code = coreError.Code()
		detail = coreError.Error()
		status = coreError.Code().Status()

		for key, value := range coreError.Fields() {
			opts = append(opts, problem.Custom(key, value))
		}
	}

	if _, problemError := problem.New(
		problem.Type(fmt.Sprintf("/api/v1/problems/%d", code)),
		problem.Status(status),
		problem.Instance(req.URL.Path),
		problem.Title(title),
		problem.Detail(detail),
	).
		Append(opts...).
		WriteTo(w); problemError != nil {
		log.Printf("Failed to encode problem '%v'. Reason: %s", err, problemError)
	}
}
