package request

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type SensorData struct {
	SensorID  string `json:"id"`
	IP        string `json:"ip"`
	EntryCode string `json:"entrycode"`
	Type      int    `json:"type"`
	Value     int    `json:"threshold"`
}

func (r SensorData) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.SensorID, validation.Required),
		validation.Field(&r.Type, validation.Min(0), validation.Max(3)),
		validation.Field(&r.Value, validation.Required),
	)
}
