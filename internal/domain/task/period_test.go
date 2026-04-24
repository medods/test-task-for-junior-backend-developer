package task

import (
	"testing"
	"time"
)

func TestPeriodNormalizeDaily(t *testing.T) {
	interval := 3
	monthDay := 12
	parity := Even
	period, err := (&Period{
		Type:          PeriodDaily,
		IntervalDays:  &interval,
		MonthDay:      &monthDay,
		SpecificDates: []time.Time{mustDate(t, "2026-04-24")},
		EvenOddType:   &parity,
	}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if period.MonthDay != nil {
		t.Fatalf("expected MonthDay to be cleared")
	}
	if period.EvenOddType != nil {
		t.Fatalf("expected EvenOddType to be cleared")
	}
	if len(period.SpecificDates) != 0 {
		t.Fatalf("expected SpecificDates to be cleared")
	}
}

func TestPeriodNormalizeSpecificDatesSortsAndDeduplicates(t *testing.T) {
	period, err := (&Period{
		Type: PeriodSpecific,
		SpecificDates: []time.Time{
			mustDateTime(t, "2026-04-25T13:00:00Z"),
			mustDateTime(t, "2026-04-24T23:59:59Z"),
			mustDateTime(t, "2026-04-25T08:00:00Z"),
		},
	}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if got, want := len(period.SpecificDates), 2; got != want {
		t.Fatalf("len(SpecificDates) = %d, want %d", got, want)
	}
	if got, want := period.SpecificDates[0], mustDate(t, "2026-04-24"); !got.Equal(want) {
		t.Fatalf("SpecificDates[0] = %s, want %s", got, want)
	}
	if got, want := period.SpecificDates[1], mustDate(t, "2026-04-25"); !got.Equal(want) {
		t.Fatalf("SpecificDates[1] = %s, want %s", got, want)
	}
}

func TestPeriodNormalizeRejectsInvalidMonthlyDay(t *testing.T) {
	monthDay := 31
	_, err := (&Period{
		Type:     PeriodMonthly,
		MonthDay: &monthDay,
	}).Normalize()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNextScheduledDateDailyUsesAnchor(t *testing.T) {
	interval := 2
	period := &Period{
		Type:         PeriodDaily,
		IntervalDays: &interval,
	}

	got := period.NextScheduledDate(
		mustDate(t, "2026-04-25"),
		mustDateTime(t, "2026-04-24T10:00:00Z"),
	)
	want := mustDate(t, "2026-04-26")

	if got == nil || !got.Equal(want) {
		t.Fatalf("NextScheduledDate() = %v, want %s", got, want)
	}
}

func TestNextScheduledDateMonthlySkipsInvalidMonth(t *testing.T) {
	monthDay := 30
	period := &Period{
		Type:     PeriodMonthly,
		MonthDay: &monthDay,
	}

	got := period.NextScheduledDate(
		mustDate(t, "2026-02-15"),
		mustDate(t, "2026-01-01"),
	)
	want := mustDate(t, "2026-03-30")

	if got == nil || !got.Equal(want) {
		t.Fatalf("NextScheduledDate() = %v, want %s", got, want)
	}
}

func TestNextScheduledDateSpecificReturnsNilWhenExhausted(t *testing.T) {
	period := &Period{
		Type: PeriodSpecific,
		SpecificDates: []time.Time{
			mustDate(t, "2026-04-20"),
			mustDate(t, "2026-04-22"),
		},
	}

	got := period.NextScheduledDate(
		mustDate(t, "2026-04-24"),
		mustDate(t, "2026-04-01"),
	)
	if got != nil {
		t.Fatalf("NextScheduledDate() = %v, want nil", got)
	}
}

func TestNextScheduledDateEvenOddFindsNextOddDay(t *testing.T) {
	parity := Odd
	period := &Period{
		Type:        PeriodEvenOdd,
		EvenOddType: &parity,
	}

	got := period.NextScheduledDate(
		mustDate(t, "2026-04-24"),
		mustDate(t, "2026-04-01"),
	)
	want := mustDate(t, "2026-04-25")

	if got == nil || !got.Equal(want) {
		t.Fatalf("NextScheduledDate() = %v, want %s", got, want)
	}
}

func mustDate(t *testing.T, raw string) time.Time {
	t.Helper()

	value, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", raw, err)
	}

	return value.UTC()
}

func mustDateTime(t *testing.T, raw string) time.Time {
	t.Helper()

	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", raw, err)
	}

	return value.UTC()
}
