package domain

import "time"

type RecurrenceType string

const (
	Daily   RecurrenceType = "daily"
	Monthly RecurrenceType = "monthly"
	Parity  RecurrenceType = "parity"
	Dates   RecurrenceType = "dates"
)

type Template struct {
	ID               int64            `json:"id" example:"1"`
	Title            string           `json:"title" example:"Title"`
	Description      string           `json:"description" example:"Description"`
	RecurrenceType   RecurrenceType   `json:"recurrence_type" example:"daily"`
	RecurrenceConfig RecurrenceConfig `json:"recurrence_config"`
	IsActive         bool             `json:"is_active" example:"true"`
	CreatedAt        time.Time        `json:"created_at" example:"2026-04-22T09:30:00Z"`
	UpdatedAt        time.Time        `json:"updated_at" example:"2026-04-22T09:30:00Z"`
}

type RecurrenceConfig struct {
	Interval int         `json:"interval,omitempty" example:"1"`
	Day      int         `json:"day,omitempty" example:"15"`
	Dates    []time.Time `json:"dates,omitempty"`
	Parity   string      `json:"parity,omitempty" example:"even"`
	Hour     int         `json:"hour" example:"10"`
	Minute   int         `json:"minute" example:"30"`
}

func (r RecurrenceType) Valid() bool {
	switch r {
	case Daily, Monthly, Parity, Dates:
		return true
	default:
		return false
	}
}
