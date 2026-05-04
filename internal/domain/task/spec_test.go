package task

import (
	"errors"
	"testing"
	"time"
)

func TestAndSpec_AggregatesErrors(t *testing.T) {
	spec := AndSpec{
		RequiredTypeSpec{},
		MonthlyDayInRangeSpec{},
	}

	cfg := RecurrenceConfig{
		Type:       "",             
		DayOfMonth: ptrInt(42),      
		StartDate:  time.Now().AddDate(0, 0, 1),
	}

	err := spec.IsSatisfiedBy(cfg)
	if err == nil {
		t.Fatal("expected joined error, got nil")
	}

	if !errors.Is(err, ErrTypeRequired) {
		t.Error("missing ErrTypeRequired in joined error")
	}
	if !errors.Is(err, ErrInvalidDayOfMonth) {
		t.Error("missing ErrInvalidDayOfMonth in joined error")
	}

	if unwrapper, ok := err.(interface{ Unwrap() []error }); ok {
		unwrapped := unwrapper.Unwrap()
		if len(unwrapped) != 2 {
			t.Errorf("expected 2 errors, got %d", len(unwrapped))
		}
	}
}

func TestDatesUniqueSpec_ValidatesDuplicates(t *testing.T) {
	spec := DatesUniqueSpec{}

	cfg := RecurrenceConfig{
		Type: RecurrenceTypeSpecificDates,
		SpecificDates: []time.Time{
			time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			time.Date(2026, 3, 20, 14, 30, 0, 0, time.UTC),
		},
	}
	if err := spec.IsSatisfiedBy(cfg); err != nil {
		t.Errorf("valid dates should pass: %v", err)
	}

	cfg.SpecificDates = []time.Time{
		time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 15, 23, 59, 59, 0, time.UTC), 
	}
	if err := spec.IsSatisfiedBy(cfg); !errors.Is(err, ErrDuplicateDate) {
		t.Errorf("expected ErrDuplicateDate, got %v", err)
	}
}