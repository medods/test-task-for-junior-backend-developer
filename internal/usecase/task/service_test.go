package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	domaintask "example.com/taskservice/internal/domain/task"
)

func TestValidateRecurrence_DailyEveryN(t *testing.T) {
	start := common.NewDate(time.Now().UTC())
	err := validateRecurrence(recurrence.DailyEveryN, recurrence.RecurrencePayload{IntervalDays: 0}, &start)
	if err == nil {
		t.Fatalf("expected error for interval 0")
	}

	if err := validateRecurrence(recurrence.DailyEveryN, recurrence.RecurrencePayload{IntervalDays: 2}, &start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRecurrence_MonthlyDay(t *testing.T) {
	start := common.NewDate(time.Now().UTC())
	if err := validateRecurrence(recurrence.MonthlyDay, recurrence.RecurrencePayload{MonthDay: 15}, &start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateRecurrence(recurrence.MonthlyDay, recurrence.RecurrencePayload{MonthDay: 31}, &start); err == nil {
		t.Fatalf("expected error for month day 31")
	}
}

func TestValidateRecurrence_SpecificDates(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	start := common.NewDate(base)
	// empty specific dates -> error
	if err := validateRecurrence(recurrence.SpecificDates, recurrence.RecurrencePayload{SpecificDates: nil}, &start); err == nil {
		t.Fatalf("expected error for empty specific dates")
	}

	// date before start -> error
	before := common.NewDate(base.Add(-24 * time.Hour))
	if err := validateRecurrence(recurrence.SpecificDates, recurrence.RecurrencePayload{SpecificDates: []common.DateType{before}}, &start); err == nil {
		t.Fatalf("expected error for specific date before start")
	}

	// duplicate dates -> error
	d := common.NewDate(base.Add(48 * time.Hour))
	if err := validateRecurrence(recurrence.SpecificDates, recurrence.RecurrencePayload{SpecificDates: []common.DateType{d, d}}, &start); err == nil {
		t.Fatalf("expected error for duplicate specific dates")
	}

	// valid
	valid := common.NewDate(base.Add(24 * time.Hour))
	if err := validateRecurrence(recurrence.SpecificDates, recurrence.RecurrencePayload{SpecificDates: []common.DateType{valid}}, &start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRecurrence_MonthParity(t *testing.T) {
	start := common.NewDate(time.Now().UTC())
	if err := validateRecurrence(recurrence.MonthParity, recurrence.RecurrencePayload{Parity: recurrence.Even}, &start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateRecurrence(recurrence.MonthParity, recurrence.RecurrencePayload{Parity: recurrence.NoneParity}, &start); err == nil {
		t.Fatalf("expected error for none parity")
	}
}

func TestValidateCreateInput_InvalidRecurrenceType(t *testing.T) {
	_, err := validateCreateInput(CreateInput{
		Title:          "t",
		Status:         domaintask.StatusNew,
		RecurrenceType: recurrence.RecurrenceType("bad"),
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidateCreateInput_DailyEveryN_RequiresInterval(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	_, err := validateCreateInput(CreateInput{
		Title:          "t",
		Status:         domaintask.StatusNew,
		RecurrenceType: recurrence.DailyEveryN,
		RecurrencePayload: recurrence.RecurrencePayload{
			IntervalDays: 0,
		},
		RecurrenceStartDate: &start,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidateCreateInput_SpecificDates_RequiresDates(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	_, err := validateCreateInput(CreateInput{
		Title:          "t",
		Status:         domaintask.StatusNew,
		RecurrenceType: recurrence.SpecificDates,
		RecurrencePayload: recurrence.RecurrencePayload{
			SpecificDates: nil,
		},
		RecurrenceStartDate: &start,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidateCreateInput_MonthParity_RequiresEvenOrOdd(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	_, err := validateCreateInput(CreateInput{
		Title:          "t",
		Status:         domaintask.StatusNew,
		RecurrenceType: recurrence.MonthParity,
		RecurrencePayload: recurrence.RecurrencePayload{
			Parity: recurrence.NoneParity,
		},
		RecurrenceStartDate: &start,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidateUpdateInput_MonthlyDay_RequiresMonthDay(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	_, err := validateUpdateInput(UpdateInput{
		Title:          "t",
		Description:    "",
		Status:         domaintask.StatusNew,
		RecurrenceType: recurrence.MonthlyDay,
		RecurrencePayload: recurrence.RecurrencePayload{
			MonthDay: 0,
		},
		RecurrenceStartDate: &start,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServicePatch_NoFieldsToUpdate(t *testing.T) {
	s := NewService(nil)
	_, err := s.Patch(context.Background(), 1, PatchInput{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
