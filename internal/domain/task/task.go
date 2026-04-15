// Package task
package task

import (
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID                  int64                        `json:"id" validate:"required"`
	Title               string                       `json:"title" validate:"required,len=128"`
	Description         string                       `json:"description" validate:"len=512"`
	Status              Status                       `json:"status" validate:"required,oneof=new in_progress done"`
	CreatedAt           time.Time                    `json:"created_at" validate:"required"`
	UpdatedAt           time.Time                    `json:"updated_at" validate:"required"`
	RecurrenceType      recurrence.RecurrenceType    `json:"recurrence_type" validate:"required"`
	RecurrencePayload   recurrence.RecurrencePayload `json:"recurrence_payload" validate:"required"`
	RecurrenceStartDate *common.DateType             `json:"recurrence_start_date"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (t Task) ActiveOn(date time.Time) bool {
	d := common.NewDate(date.UTC()).Time
	if t.RecurrenceType == recurrence.NoneRecurrence {
		if t.RecurrenceStartDate == nil {
			return false
		}
		start := common.NewDate(t.RecurrenceStartDate.Time.UTC()).Time
		return start.Equal(d)
	}

	if t.RecurrenceStartDate == nil {
		return false
	}
	start := common.NewDate(t.RecurrenceStartDate.Time.UTC()).Time
	if d.Before(start) {
		return false
	}

	if t.RecurrenceType == recurrence.DailyEveryN {
		if t.RecurrencePayload.IntervalDays <= 0 || t.RecurrenceStartDate == nil {
			return false
		}
		diff := t.RecurrenceStartDate.Sub(date)

		if diff >= 0 && diff%time.Duration(t.RecurrencePayload.IntervalDays) == 0 {
			return true
		}
		return false
	}

	switch t.RecurrenceType {
	case recurrence.DailyEveryN:
		n := t.RecurrencePayload.IntervalDays
		if n <= 0 {
			return false
		}
		days := int(d.Sub(start).Hours() / 24)
		return days%n == 0
	case recurrence.MonthlyDay:
		md := t.RecurrencePayload.MonthDay
		if md < 1 || md > 30 {
			return false
		}
		return d.Day() == md
	case recurrence.SpecificDates:
		for _, sd := range t.RecurrencePayload.SpecificDates {
			if common.NewDate(sd.Time.UTC()).Equal(d) {
				return true
			}
		}
		return false
	case recurrence.MonthParity:
		p := t.RecurrencePayload.Parity
		if p == recurrence.NoneParity {
			return false
		}
		if p == recurrence.Even {
			return d.Day()%2 == 0
		}
		return d.Day()%2 != 0
	default:
		return false
	}
}
