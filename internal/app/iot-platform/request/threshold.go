package request

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type SensorThreshold struct {
	Type      int `json:"type"`
	Threshold int `json:"threshold"`
}

func (r SensorThreshold) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Type, validation.Required, validation.Min(0), validation.Max(1)),
		validation.Field(&r.Threshold, validation.Required),
	)
}
