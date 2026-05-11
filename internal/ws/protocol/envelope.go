package protocol

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
)

type Envelope struct {
	Type    string          `json:"type" validate:"required"`
	Payload json.RawMessage `json:"payload"`
}

func (e *Envelope) Validate() error {
	if e == nil {
		return errors.New("envelope is nil")
	}
	v := validator.New()
	return v.Struct(e)
}

func (e Envelope) UnmarshalPayload(v any) error {
	return json.Unmarshal(e.Payload, v)
}

type HandlerFunc func(ctx context.Context, payload json.RawMessage) (interface{}, error)
