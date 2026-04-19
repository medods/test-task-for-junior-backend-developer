package task

import (
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func dates(ds ...time.Time) []time.Time { return ds }

func TestGenerateDates_UnknownType(t *testing.T) {
	_, err := generateDates(RecurrenceInput{Type: "weekly"})
	if err == nil {
		t.Fatal("expected error for unknown recurrence type, got nil")
	}
}

func TestGenerateDates_SpecificDates_ReturnedAsIs(t *testing.T) {
	input := dates(date(2024, 3, 1), date(2024, 3, 15))
	result, err := generateDates(RecurrenceInput{
		Type:  string(SpecificDates),
		Dates: input,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, input)
}

func TestGenerateDaily_DefaultStep(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(Daily),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 5),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 1),
		date(2024, 1, 2),
		date(2024, 1, 3),
		date(2024, 1, 4),
		date(2024, 1, 5),
	))
}

func TestGenerateDaily_EveryNDays(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Daily),
		StartDate:  date(2024, 1, 1),
		EndDate:    date(2024, 1, 10),
		EveryNDays: ptr(3),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 1),
		date(2024, 1, 4),
		date(2024, 1, 7),
		date(2024, 1, 10),
	))
}

func TestGenerateDaily_ZeroNDaysFallsBackToDefault(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Daily),
		StartDate:  date(2024, 1, 1),
		EndDate:    date(2024, 1, 3),
		EveryNDays: ptr(0),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 1),
		date(2024, 1, 2),
		date(2024, 1, 3),
	))
}

func TestGenerateDaily_StartEqualsEnd(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(Daily),
		StartDate: date(2024, 6, 15),
		EndDate:   date(2024, 6, 15),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(date(2024, 6, 15)))
}

func TestGenerateDaily_StartAfterEnd_ReturnsEmpty(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(Daily),
		StartDate: date(2024, 6, 20),
		EndDate:   date(2024, 6, 15),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}
}

func TestGenerateDaily_AcrossMonthBoundary(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(Daily),
		StartDate: date(2024, 1, 30),
		EndDate:   date(2024, 2, 2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 30),
		date(2024, 1, 31),
		date(2024, 2, 1),
		date(2024, 2, 2),
	))
}

func TestGenerateMonthly_TypicalCase(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 1, 15),
		EndDate:    date(2024, 4, 15),
		DayOfMonth: ptr(15),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 15),
		date(2024, 2, 15),
		date(2024, 3, 15),
		date(2024, 4, 15),
	))
}

func TestGenerateMonthly_StartBeforeTargetDay(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 1, 10),
		EndDate:    date(2024, 3, 1),
		DayOfMonth: ptr(20),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 20),
		date(2024, 2, 20),
	))
}

func TestGenerateMonthly_StartAfterTargetDay(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 1, 25),
		EndDate:    date(2024, 4, 1),
		DayOfMonth: ptr(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 2, 10),
		date(2024, 3, 10),
	))
}

func TestGenerateMonthly_Day31ClampsToLastDay(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 1, 1),
		EndDate:    date(2024, 5, 1),
		DayOfMonth: ptr(31),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 31),
		date(2024, 2, 29),
		date(2024, 3, 31),
		date(2024, 4, 30),
	))
}

func TestGenerateMonthly_Day29LeapYear(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 2, 1),
		EndDate:    date(2024, 3, 31),
		DayOfMonth: ptr(29),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 2, 29),
		date(2024, 3, 29),
	))
}

func TestGenerateMonthly_Day29NonLeapYear(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2023, 2, 1),
		EndDate:    date(2023, 3, 31),
		DayOfMonth: ptr(29),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2023, 2, 28),
		date(2023, 3, 29),
	))
}

func TestGenerateMonthly_ExactlyOnStartDate(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:       string(Monthly),
		StartDate:  date(2024, 3, 10),
		EndDate:    date(2024, 3, 31),
		DayOfMonth: ptr(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(date(2024, 3, 10)))
}

func TestGenerateEvenOdd_EvenDays(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(EvenOdd),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 6),
		Parity:    ptr(string(Even)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 2),
		date(2024, 1, 4),
		date(2024, 1, 6),
	))
}

func TestGenerateEvenOdd_OddDays(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(EvenOdd),
		StartDate: date(2024, 1, 1),
		EndDate:   date(2024, 1, 6),
		Parity:    ptr(string(Odd)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 1),
		date(2024, 1, 3),
		date(2024, 1, 5),
	))
}

func TestGenerateEvenOdd_AcrossMonthBoundary(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(EvenOdd),
		StartDate: date(2024, 1, 30),
		EndDate:   date(2024, 2, 3),
		Parity:    ptr(string(Odd)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(
		date(2024, 1, 31),
		date(2024, 2, 1),
		date(2024, 2, 3),
	))
}

func TestGenerateEvenOdd_SingleDayRange_Matches(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(EvenOdd),
		StartDate: date(2024, 1, 4),
		EndDate:   date(2024, 1, 4),
		Parity:    ptr(string(Even)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDates(t, result, dates(date(2024, 1, 4)))
}

func TestGenerateEvenOdd_SingleDayRange_NoMatch(t *testing.T) {
	result, err := generateDates(RecurrenceInput{
		Type:      string(EvenOdd),
		StartDate: date(2024, 1, 4),
		EndDate:   date(2024, 1, 4),
		Parity:    ptr(string(Odd)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestAdjustDay_NormalDay(t *testing.T) {
	got := adjustDay(date(2024, 3, 1), 15)
	want := date(2024, 3, 15)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestAdjustDay_DayExceedsMonthLength(t *testing.T) {
	got := adjustDay(date(2024, 2, 1), 31)
	want := date(2024, 2, 29)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestAdjustDay_LastDayOfMonth(t *testing.T) {
	got := adjustDay(date(2024, 4, 1), 30)
	want := date(2024, 4, 30)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFirstMonthDayOnOrAfter_SameDay(t *testing.T) {
	got := firstMonthDayOnOrAfter(date(2024, 3, 15), 15)
	want := date(2024, 3, 15)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFirstMonthDayOnOrAfter_TargetDayAhead(t *testing.T) {
	got := firstMonthDayOnOrAfter(date(2024, 3, 5), 20)
	want := date(2024, 3, 20)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFirstMonthDayOnOrAfter_TargetDayPassed_SkipsToNextMonth(t *testing.T) {
	got := firstMonthDayOnOrAfter(date(2024, 3, 25), 10)
	want := date(2024, 4, 10)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFirstMonthDayOnOrAfter_TargetDayExceedsFebruary(t *testing.T) {
	got := firstMonthDayOnOrAfter(date(2023, 2, 1), 30)
	want := date(2023, 2, 28)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func assertDates(t *testing.T, got, want []time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length mismatch: want %d dates %v, got %d dates %v", len(want), want, len(got), got)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("index %d: want %v, got %v", i, want[i], got[i])
		}
	}
}
