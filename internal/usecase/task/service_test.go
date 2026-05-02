package task

import (
	"errors"
	"testing"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func intPtr(v int) *int {
	return &v
}

func parityPtr(v taskdomain.MonthDayParity) *taskdomain.MonthDayParity {
	return &v
}

func TestValidateRecurrence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     taskdomain.Recurrence
		assertion func(t *testing.T, got taskdomain.Recurrence, err error)
	}{
		{
			name:  "defaults to none when omitted",
			input: taskdomain.Recurrence{},
			assertion: func(t *testing.T, got taskdomain.Recurrence, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Type != taskdomain.RecurrenceTypeNone {
					t.Fatalf("unexpected type: %q", got.Type)
				}
			},
		},
		{
			name: "daily requires every_n_days",
			input: taskdomain.Recurrence{
				Type: taskdomain.RecurrenceTypeDaily,
			},
			assertion: func(t *testing.T, _ taskdomain.Recurrence, err error) {
				t.Helper()
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("expected ErrInvalidInput, got: %v", err)
				}
			},
		},
		{
			name: "monthly day out of range",
			input: taskdomain.Recurrence{
				Type:       taskdomain.RecurrenceTypeMonthly,
				DayOfMonth: intPtr(31),
			},
			assertion: func(t *testing.T, _ taskdomain.Recurrence, err error) {
				t.Helper()
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("expected ErrInvalidInput, got: %v", err)
				}
			},
		},
		{
			name: "specific dates are normalized",
			input: taskdomain.Recurrence{
				Type:  taskdomain.RecurrenceTypeSpecificDates,
				Dates: []string{" 2026-05-03 ", "2026-05-01", "2026-05-01"},
			},
			assertion: func(t *testing.T, got taskdomain.Recurrence, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(got.Dates) != 2 {
					t.Fatalf("unexpected dates count: %d", len(got.Dates))
				}
				if got.Dates[0] != "2026-05-01" || got.Dates[1] != "2026-05-03" {
					t.Fatalf("unexpected normalized dates: %#v", got.Dates)
				}
			},
		},
		{
			name: "month parity validates parity value",
			input: taskdomain.Recurrence{
				Type:   taskdomain.RecurrenceTypeMonthParity,
				Parity: parityPtr(taskdomain.MonthDayParityEven),
			},
			assertion: func(t *testing.T, got taskdomain.Recurrence, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Parity == nil || *got.Parity != taskdomain.MonthDayParityEven {
					t.Fatalf("unexpected parity: %#v", got.Parity)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := validateRecurrence(tc.input)
			tc.assertion(t, got, err)
		})
	}
}

func TestValidateCreateInput_WithDefaultRecurrence(t *testing.T) {
	t.Parallel()

	input := CreateInput{
		Title:  "Call patients",
		Status: taskdomain.StatusNew,
	}

	got, err := validateCreateInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Recurrence.Type != taskdomain.RecurrenceTypeNone {
		t.Fatalf("expected default recurrence type none, got %q", got.Recurrence.Type)
	}
}
