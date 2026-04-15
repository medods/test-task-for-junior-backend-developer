package schedule

import (
	"testing"
	"time"
)

func TestDailyIntervalOccurrences(t *testing.T) {
	t.Parallel()

	startDate := mustDate(t, "2026-04-15")
	windowStart := mustDate(t, "2026-04-15")
	windowEnd := mustDate(t, "2026-04-21")

	rule := DailyIntervalRule{EveryDays: 2}
	got := rule.Occurrences(startDate, nil, windowStart, windowEnd)

	assertDates(t, got, "2026-04-15", "2026-04-17", "2026-04-19", "2026-04-21")
}

func TestMonthlyDayOccurrencesSkipMissingDates(t *testing.T) {
	t.Parallel()

	startDate := mustDate(t, "2026-01-01")
	windowStart := mustDate(t, "2026-02-01")
	windowEnd := mustDate(t, "2026-03-31")

	rule := MonthlyDayRule{DayOfMonth: 30}
	got := rule.Occurrences(startDate, nil, windowStart, windowEnd)

	assertDates(t, got, "2026-03-30")
}

func TestSpecificDatesRuleRejectsDuplicates(t *testing.T) {
	t.Parallel()

	duplicateDate := mustDate(t, "2026-05-01")
	rule := SpecificDatesRule{Dates: []time.Time{duplicateDate, duplicateDate}}

	if err := rule.Validate(); err == nil {
		t.Fatalf("expected duplicate dates validation error")
	}
}

func TestMonthParityOccurrences(t *testing.T) {
	t.Parallel()

	startDate := mustDate(t, "2026-04-01")
	windowStart := mustDate(t, "2026-04-01")
	windowEnd := mustDate(t, "2026-04-07")

	rule := MonthParityRule{Parity: MonthParityOdd}
	got := rule.Occurrences(startDate, nil, windowStart, windowEnd)

	assertDates(t, got, "2026-04-01", "2026-04-03", "2026-04-05", "2026-04-07")
}

func TestEncodeDecodeRuleRoundTripSpecificDates(t *testing.T) {
	t.Parallel()

	rule := SpecificDatesRule{
		Dates: []time.Time{
			mustDate(t, "2026-05-01"),
			mustDate(t, "2026-05-15"),
		},
	}

	ruleType, payload, err := EncodeRule(rule)
	if err != nil {
		t.Fatalf("encode rule: %v", err)
	}

	decoded, err := DecodeRule(ruleType, payload)
	if err != nil {
		t.Fatalf("decode rule: %v", err)
	}

	typed, ok := decoded.(SpecificDatesRule)
	if !ok {
		t.Fatalf("unexpected decoded rule type: %T", decoded)
	}

	assertDates(t, typed.Dates, "2026-05-01", "2026-05-15")
}

func mustDate(t *testing.T, raw string) time.Time {
	t.Helper()

	value, err := ParseDate(raw)
	if err != nil {
		t.Fatalf("parse date %q: %v", raw, err)
	}

	return value
}

func assertDates(t *testing.T, got []time.Time, expected ...string) {
	t.Helper()

	if len(got) != len(expected) {
		t.Fatalf("unexpected dates length: got %d want %d", len(got), len(expected))
	}

	for index, raw := range expected {
		want := mustDate(t, raw)
		if !got[index].Equal(want) {
			t.Fatalf("unexpected date at index %d: got %s want %s", index, FormatDate(got[index]), raw)
		}
	}
}
