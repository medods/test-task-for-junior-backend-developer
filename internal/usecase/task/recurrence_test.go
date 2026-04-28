package task

import (
	"errors"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestValidateAndNormalizeRecurrenceSpecificDates(t *testing.T) {
	first := mustDateTime(2026, time.May, 10, 14, 30, time.FixedZone("UTC+3", 3*60*60))
	second := mustDateTime(2026, time.May, 1, 8, 15, time.FixedZone("UTC-5", -5*60*60))

	recurrence := &taskdomain.Recurrence{
		Type:  taskdomain.RecurrenceSpecificDates,
		Dates: []time.Time{first, second},
	}

	got, err := validateAndNormalizeRecurrence(recurrence)
	if err != nil {
		t.Fatalf("validateAndNormalizeRecurrence() error = %v", err)
	}

	want := []time.Time{
		mustDate(t, "2026-05-01"),
		mustDate(t, "2026-05-10"),
	}

	assertDateStringsEqual(t, got.Dates, want)
}

func TestValidateAndNormalizeRecurrenceRejectsDuplicateSpecificDates(t *testing.T) {
	recurrence := &taskdomain.Recurrence{
		Type: taskdomain.RecurrenceSpecificDates,
		Dates: []time.Time{
			mustDateTime(2026, time.May, 10, 10, 0, time.UTC),
			mustDateTime(2026, time.May, 10, 18, 0, time.UTC),
		},
	}

	_, err := validateAndNormalizeRecurrence(recurrence)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("validateAndNormalizeRecurrence() error = %v, want ErrInvalidInput", err)
	}
}

func TestValidateAndNormalizeRecurrenceRejectsMonthlyWithoutStartDate(t *testing.T) {
	recurrence := &taskdomain.Recurrence{
		Type:       taskdomain.RecurrenceMonthly,
		DayOfMonth: 15,
	}

	_, err := validateAndNormalizeRecurrence(recurrence)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("validateAndNormalizeRecurrence() error = %v, want ErrInvalidInput", err)
	}
}

func TestGenerateScheduledDatesDaily(t *testing.T) {
	startDate := mustDate(t, "2026-04-28")
	recurrence := &taskdomain.Recurrence{
		Type:       taskdomain.RecurrenceDaily,
		EveryNDays: 3,
		StartDate:  &startDate,
	}

	got := generateScheduledDates(recurrence, mustDate(t, "2026-04-28"), 10)
	want := []time.Time{
		mustDate(t, "2026-04-28"),
		mustDate(t, "2026-05-01"),
		mustDate(t, "2026-05-04"),
		mustDate(t, "2026-05-07"),
	}

	assertDateStringsEqual(t, got, want)
}

func TestGenerateScheduledDatesMonthlyClampsShortMonth(t *testing.T) {
	startDate := mustDate(t, "2026-01-01")
	recurrence := &taskdomain.Recurrence{
		Type:       taskdomain.RecurrenceMonthly,
		DayOfMonth: 30,
		StartDate:  &startDate,
	}

	got := generateScheduledDates(recurrence, mustDate(t, "2026-01-15"), 80)
	want := []time.Time{
		mustDate(t, "2026-01-30"),
		mustDate(t, "2026-02-28"),
		mustDate(t, "2026-03-30"),
	}

	assertDateStringsEqual(t, got, want)
}

func TestGenerateScheduledDatesSpecificDatesFiltersRange(t *testing.T) {
	recurrence := &taskdomain.Recurrence{
		Type: taskdomain.RecurrenceSpecificDates,
		Dates: []time.Time{
			mustDate(t, "2026-04-20"),
			mustDate(t, "2026-04-28"),
			mustDate(t, "2026-05-10"),
			mustDate(t, "2026-06-01"),
		},
	}

	got := generateScheduledDates(recurrence, mustDate(t, "2026-04-28"), 20)
	want := []time.Time{
		mustDate(t, "2026-04-28"),
		mustDate(t, "2026-05-10"),
	}

	assertDateStringsEqual(t, got, want)
}

func TestGenerateScheduledDatesOddEvenRespectsStartDate(t *testing.T) {
	startDate := mustDate(t, "2026-04-29")

	testCases := []struct {
		name      string
		recType   taskdomain.RecurrenceType
		wantDates []time.Time
	}{
		{
			name:    "odd days",
			recType: taskdomain.RecurrenceOddDays,
			wantDates: []time.Time{
				mustDate(t, "2026-04-29"),
				mustDate(t, "2026-05-01"),
			},
		},
		{
			name:    "even days",
			recType: taskdomain.RecurrenceEvenDays,
			wantDates: []time.Time{
				mustDate(t, "2026-04-30"),
				mustDate(t, "2026-05-02"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recurrence := &taskdomain.Recurrence{
				Type:      tc.recType,
				StartDate: &startDate,
			}

			got := generateScheduledDates(recurrence, mustDate(t, "2026-04-28"), 5)
			assertDateStringsEqual(t, got, tc.wantDates)
		})
	}
}
