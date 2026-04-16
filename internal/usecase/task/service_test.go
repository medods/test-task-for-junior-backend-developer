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
