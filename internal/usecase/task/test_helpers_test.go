package task

import (
	"reflect"
	"testing"
	"time"
)

func mustDate(t *testing.T, raw string) time.Time {
	t.Helper()

	value, err := time.ParseInLocation(time.DateOnly, raw, time.UTC)
	if err != nil {
		t.Fatalf("parse date %q: %v", raw, err)
	}

	return value
}

func mustDateTime(
	year int,
	month time.Month,
	day int,
	hour int,
	minute int,
	loc *time.Location,
) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, loc)
}

func dateStrings(values []time.Time) []string {
	result := make([]string, 0, len(values))

	for _, value := range values {
		result = append(result, dateOnly(value).Format(time.DateOnly))
	}

	return result
}

func assertDateStringsEqual(t *testing.T, got []time.Time, want []time.Time) {
	t.Helper()

	gotStrings := dateStrings(got)
	wantStrings := dateStrings(want)

	if !reflect.DeepEqual(gotStrings, wantStrings) {
		t.Fatalf("dates mismatch: got %v, want %v", gotStrings, wantStrings)
	}
}
