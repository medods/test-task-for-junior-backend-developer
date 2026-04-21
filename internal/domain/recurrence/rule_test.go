package recurrence

import (
	"testing"
	"time"

	"example.com/taskservice/internal/shared/dateutil"
)

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := dateutil.Parse(value)
	if err != nil {
		t.Fatalf("parse date %s: %v", value, err)
	}

	return parsed
}

func TestOccurrencesBetweenEveryNDays(t *testing.T) {
	interval := 2
	rule := Rule{
		Active:       true,
		ScheduleType: ScheduleTypeEveryNDays,
		StartsOn:     mustDate(t, "2026-04-20"),
		EveryNDays:   &interval,
	}

	occurrences := rule.OccurrencesBetween(mustDate(t, "2026-04-20"), mustDate(t, "2026-04-26"))
	if len(occurrences) != 4 {
		t.Fatalf("expected 4 occurrences, got %d", len(occurrences))
	}

	expected := []string{"2026-04-20", "2026-04-22", "2026-04-24", "2026-04-26"}
	for i, occurrence := range occurrences {
		if got := dateutil.Format(occurrence); got != expected[i] {
			t.Fatalf("occurrence %d: expected %s, got %s", i, expected[i], got)
		}
	}
}

func TestOccurrencesBetweenMonthlyDay(t *testing.T) {
	day := 15
	rule := Rule{
		Active:       true,
		ScheduleType: ScheduleTypeMonthlyDay,
		StartsOn:     mustDate(t, "2026-04-01"),
		DayOfMonth:   &day,
	}

	occurrences := rule.OccurrencesBetween(mustDate(t, "2026-04-10"), mustDate(t, "2026-06-20"))
	expected := []string{"2026-04-15", "2026-05-15", "2026-06-15"}
	if len(occurrences) != len(expected) {
		t.Fatalf("expected %d occurrences, got %d", len(expected), len(occurrences))
	}

	for i, occurrence := range occurrences {
		if got := dateutil.Format(occurrence); got != expected[i] {
			t.Fatalf("occurrence %d: expected %s, got %s", i, expected[i], got)
		}
	}
}

func TestOccurrencesBetweenSpecificDates(t *testing.T) {
	rule := Rule{
		Active:        true,
		ScheduleType:  ScheduleTypeSpecificDates,
		StartsOn:      mustDate(t, "2026-04-20"),
		SpecificDates: []time.Time{mustDate(t, "2026-04-22"), mustDate(t, "2026-05-01"), mustDate(t, "2026-04-20")},
	}

	occurrences := rule.OccurrencesBetween(mustDate(t, "2026-04-21"), mustDate(t, "2026-04-30"))
	if len(occurrences) != 1 {
		t.Fatalf("expected 1 occurrence, got %d", len(occurrences))
	}
	if got := dateutil.Format(occurrences[0]); got != "2026-04-22" {
		t.Fatalf("expected 2026-04-22, got %s", got)
	}
}

func TestOccurrencesBetweenParity(t *testing.T) {
	parity := DayParityEven
	rule := Rule{
		Active:       true,
		ScheduleType: ScheduleTypeMonthParity,
		StartsOn:     mustDate(t, "2026-04-20"),
		DayParity:    &parity,
	}

	occurrences := rule.OccurrencesBetween(mustDate(t, "2026-04-20"), mustDate(t, "2026-04-25"))
	expected := []string{"2026-04-20", "2026-04-22", "2026-04-24"}
	if len(occurrences) != len(expected) {
		t.Fatalf("expected %d occurrences, got %d", len(expected), len(occurrences))
	}

	for i, occurrence := range occurrences {
		if got := dateutil.Format(occurrence); got != expected[i] {
			t.Fatalf("occurrence %d: expected %s, got %s", i, expected[i], got)
		}
	}
}
