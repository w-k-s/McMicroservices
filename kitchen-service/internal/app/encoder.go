package app

import (
	"encoding/json"
	"io"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"
)

type Encoder interface {
	Encode(w io.Writer, v interface{}) error
	MustEncode(w io.Writer, any interface{})
}

type Decoder interface {
	Decode(r io.Reader, v interface{}) error
	MustDecode(r io.Reader, v interface{})
}

type JSONEncoder struct{}

func (je JSONEncoder) Encode(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	return encoder.Encode(v)
}

func (je JSONEncoder) MustEncode(w io.Writer, v interface{}) {
	if err := je.Encode(w, v); err != nil {
		log.Fatalf("failed to encode json. Reason: %s", err)
	}
}

type JSONDecoder struct{}

func (jd JSONDecoder) Decode(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}

func (jd JSONDecoder) MustDecode(r io.Reader, v interface{}) {
	if err := jd.Decode(r, v); err != nil {
		log.Fatalf("failed to decode json. Reason: %s", err)
	}
}
