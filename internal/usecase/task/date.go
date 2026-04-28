package task

import "time"

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func clampedDateInMonth(year int, month time.Month, day int) time.Time {
	lastDay := daysInMonth(year, month)

	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
	firstDayOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := firstDayOfNextMonth.AddDate(0, 0, -1)

	return lastDayOfMonth.Day()
}
