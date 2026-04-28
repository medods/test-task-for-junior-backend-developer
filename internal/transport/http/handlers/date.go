package handlers

import (
	"time"
)

const dateLayout = "2006-01-02"

func parseDate(raw string) (time.Time, error) {
	parsed, err := time.ParseInLocation(dateLayout, raw, time.UTC)
	if err != nil {
		return time.Time{}, err
	}

	year, month, day := parsed.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
}

func parseDatePtr(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}

	parsed, err := parseDate(*raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseDateSlice(values []string) ([]time.Time, error) {
	if len(values) == 0 {
		return nil, nil
	}

	result := make([]time.Time, 0, len(values))

	for _, value := range values {
		parsed, err := parseDate(value)
		if err != nil {
			return nil, err
		}

		result = append(result, parsed)
	}

	return result, nil
}

func formatDate(value time.Time) string {
	return value.UTC().Format(dateLayout)
}

func formatDatePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := formatDate(*value)

	return &formatted
}

func formatDateSlice(values []time.Time) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))

	for _, value := range values {
		result = append(result, formatDate(value))
	}

	return result
}
