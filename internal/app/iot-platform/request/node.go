package request

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type NodeRequest struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	EntryCode string `json:"entry_code"`
	Type      int    `json:"type"`
}

type NodeUpdate struct {
	ID    string `json:"id"`
	State int    `json:"state"`
}

func (r NodeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required),
		validation.Field(&r.IP, validation.Required),
		validation.Field(&r.EntryCode, validation.Required),
	)
}

func (r NodeUpdate) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required),
		validation.Field(&r.State, validation.Min(0), validation.Max(1)),
	)
}
