package task

import (
	"errors"
	"testing"
	"time"
)

func TestValidateRecurrenceInput_EndDateBeforeStart(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:       string(Daily),
		StartDate:  date(2024, 6, 10),
		EndDate:    date(2024, 6, 1),
		EveryNDays: ptr(1),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_Daily_MissingNDays(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:      string(Daily),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 31),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_Daily_Valid(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:       string(Daily),
		StartDate:  date(2024, 1, 1),
		EndDate:    date(2024, 1, 31),
		EveryNDays: ptr(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRecurrenceInput_Monthly_MissingDayOfMonth(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:      string(Monthly),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 6, 30),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_Monthly_DayOfMonthTooLow(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 1, 1),
		EndDate:    date(2024, 6, 30),
		DayOfMonth: ptr(0),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_Monthly_DayOfMonthTooHigh(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 1, 1),
		EndDate:    date(2024, 6, 30),
		DayOfMonth: ptr(32),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_Monthly_Valid(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 1, 1),
		EndDate:    date(2024, 6, 30),
		DayOfMonth: ptr(15),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRecurrenceInput_EvenOdd_MissingParity(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:      string(EvenOdd),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 31),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_EvenOdd_InvalidParity(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:      string(EvenOdd),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 31),
		Parity:    ptr("weekday"),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_EvenOdd_Valid(t *testing.T) {
	for _, p := range []string{string(Even), string(Odd)} {
		t.Run(p, func(t *testing.T) {
			_, err := validateRecurrenceInput(RecurrenceInput{
				Type:      string(EvenOdd),
				StartDate: date(2024, 1, 1),
				EndDate:   date(2024, 1, 31),
				Parity:    ptr(p),
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRecurrenceInput_SpecificDates_NoDates(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:      string(SpecificDates),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 31),
	})
	assertError(t, err, ErrInvalidRecurrence)
}

func TestValidateRecurrenceInput_SpecificDates_Valid(t *testing.T) {
	_, err := validateRecurrenceInput(RecurrenceInput{
		Type:      string(SpecificDates),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 31),
		Dates:     []time.Time{date(2024, 1, 10), date(2024, 1, 20)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error, target error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error wrapping %v, got nil", target)
	}
	if !errors.Is(err, target) {
		t.Fatalf("expected error wrapping %v, got: %v", target, err)
	}
}
