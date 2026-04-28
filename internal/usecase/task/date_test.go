package task

import (
	"testing"
	"time"
)

func TestDateOnly(t *testing.T) {
	value := mustDateTime(2026, time.April, 28, 15, 45, time.FixedZone("UTC+3", 3*60*60))

	got := dateOnly(value)
	want := mustDate(t, "2026-04-28")

	if !got.Equal(want) {
		t.Fatalf("dateOnly() = %v, want %v", got, want)
	}

	if got.Location() != time.UTC {
		t.Fatalf("dateOnly() location = %v, want UTC", got.Location())
	}
}

func TestClampedDateInMonth(t *testing.T) {
	testCases := []struct {
		name  string
		year  int
		month time.Month
		day   int
		want  time.Time
	}{
		{
			name:  "keeps existing day",
			year:  2026,
			month: time.April,
			day:   15,
			want:  mustDate(t, "2026-04-15"),
		},
		{
			name:  "clamps february non leap year",
			year:  2026,
			month: time.February,
			day:   30,
			want:  mustDate(t, "2026-02-28"),
		},
		{
			name:  "clamps february leap year",
			year:  2028,
			month: time.February,
			day:   30,
			want:  mustDate(t, "2028-02-29"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := clampedDateInMonth(tc.year, tc.month, tc.day)

			if !got.Equal(tc.want) {
				t.Fatalf("clampedDateInMonth() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDaysInMonth(t *testing.T) {
	testCases := []struct {
		name  string
		year  int
		month time.Month
		want  int
	}{
		{name: "april", year: 2026, month: time.April, want: 30},
		{name: "may", year: 2026, month: time.May, want: 31},
		{name: "february non leap", year: 2026, month: time.February, want: 28},
		{name: "february leap", year: 2028, month: time.February, want: 29},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := daysInMonth(tc.year, tc.month); got != tc.want {
				t.Fatalf("daysInMonth() = %d, want %d", got, tc.want)
			}
		})
	}
}
