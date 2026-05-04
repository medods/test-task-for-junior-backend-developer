package task

import (
	"testing"
	"testing/quick"
	"time"
)

func TestCalculateEffectiveStart_Properties(t *testing.T) {
	property := func(startOffsetDays int, interval int, nowOffsetDays int) bool {
		if interval <= 0 || interval > 365 {
			return true 
		}

		base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

		start := base.AddDate(0, 0, startOffsetDays%361-180)

		now := base.AddDate(0, 0, nowOffsetDays%11)

		cfg := RecurrenceConfig{
			Type:      RecurrenceTypeDaily,
			Interval:  &interval,
			StartDate: start,
		}

		effective := cfg.CalculateEffectiveStart(now)

		if effective.Before(now) {
			return false
		}

		if effective.Before(start) {
			return false
		}

		diffDays := int(effective.Sub(start).Hours() / 24)
		if diffDays%interval != 0 {
			return false
		}

		if effective.Hour() != 0 || effective.Minute() != 0 || effective.Second() != 0 {
			return false
		}

		return true
	}

	if err := quick.Check(property, nil); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}