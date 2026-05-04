package task

import (
	"fmt"
	"time"
)

func IsValidDayInMonth(year, month, day int) bool {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}

	candidate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return candidate.Day() == day
}

func NormalizeToUTCStartOfDay(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

func ApplyTimezone(cfg *RecurrenceConfig) error {
	if cfg == nil {
		return fmt.Errorf("apply timezone: config is nil")
	}

	loc := time.UTC
	if cfg.Timezone != "" {
		var err error
		loc, err = time.LoadLocation(cfg.Timezone)
		if err != nil {
			return fmt.Errorf("apply timezone: %w", err)
		}
	}

	convertToUTC := func(t time.Time) time.Time {
		y, m, d := t.Date()
		h, min, s := t.Clock()
		local := time.Date(y, m, d, h, min, s, t.Nanosecond(), loc)
		return local.UTC()
	}

	cfg.StartDate = convertToUTC(cfg.StartDate)

	if cfg.EndDate != nil {
		end := convertToUTC(*cfg.EndDate)
		cfg.EndDate = &end
	}

	if len(cfg.SpecificDates) > 0 {
		for i, d := range cfg.SpecificDates {
			cfg.SpecificDates[i] = convertToUTC(d)
		}
	}

	return nil
}

func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func (c *RecurrenceConfig) resolveLocation() *time.Location {
	if c.Timezone != "" {
		if loc, err := time.LoadLocation(c.Timezone); err == nil {
			return loc
		}
	}

	if loc := c.StartDate.Location(); loc != nil {
		return loc
	}

	return time.UTC
}