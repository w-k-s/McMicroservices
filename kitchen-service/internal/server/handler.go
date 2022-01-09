package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
		h.MustEncodeProblem(w, req, k.NewError(k.ErrUnmarshalling, "Failed to parse request", err))
		return false
	}
	return true
}

func (h Handler) MustEncodeProblem(w http.ResponseWriter, req *http.Request, err error) {

	log.Printf("Error: %s", err.Error())

	title := k.ErrUnknown.Name()
	code := k.ErrUnknown
	detail := err.Error()
	opts := []problem.Option{}
	status := k.ErrUnknown.Status()

	if coreError, ok := err.(k.Error); ok {
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

func (h Handler) MustMarshal(result []byte, err error) []byte{
	if err != nil {
		log.Fatalf("Marshal to json failed. Reason: %s", err)
	}
	return result
}