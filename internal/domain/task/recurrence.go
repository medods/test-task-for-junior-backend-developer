package task

import (
	"fmt"
	"time"
)

type Recurrence interface {
	RecurrenceType() string
	Valid() error
	NextOccurrence(from time.Time) *time.Time
}

type DailyRecurrence struct {
	Interval int `json:"interval"`
}

func (r DailyRecurrence) RecurrenceType() string { return "daily" }

func (r DailyRecurrence) Valid() error {
	if r.Interval <= 0 {
		return fmt.Errorf("daily interval must be greater than 0")
	}
	return nil
}

func (r DailyRecurrence) NextOccurrence(from time.Time) *time.Time {
	next := from.AddDate(0, 0, r.Interval)
	next = truncateToDate(next)
	return &next
}

type MonthlyRecurrence struct {
	DayOfMonth int `json:"day_of_month"`
}

func (r MonthlyRecurrence) RecurrenceType() string { return "monthly" }

func (r MonthlyRecurrence) Valid() error {
	if r.DayOfMonth < 1 || r.DayOfMonth > 30 {
		return fmt.Errorf("day_of_month must be between 1 and 30")
	}
	return nil
}

func (r MonthlyRecurrence) NextOccurrence(from time.Time) *time.Time {
	from = truncateToDate(from)
	candidate := time.Date(from.Year(), from.Month(), r.DayOfMonth, 0, 0, 0, 0, from.Location())
	if !candidate.After(from) {
		candidate = time.Date(from.Year(), from.Month()+1, r.DayOfMonth, 0, 0, 0, 0, from.Location())
	}
	for candidate.Day() != r.DayOfMonth {
		candidate = time.Date(candidate.Year(), candidate.Month()+1, r.DayOfMonth, 0, 0, 0, 0, candidate.Location())
	}
	return &candidate
}

type WeeklyRecurrence struct {
	WeekDays []int `json:"week_days"`
}

func (r WeeklyRecurrence) RecurrenceType() string { return "weekly" }

func (r WeeklyRecurrence) Valid() error {
	if len(r.WeekDays) == 0 {
		return fmt.Errorf("week_days must not be empty")
	}
	seen := make(map[int]bool)
	for _, d := range r.WeekDays {
		if d < 1 || d > 7 {
			return fmt.Errorf("week_days values must be between 1 (mon) and 7 (sun)")
		}
		if seen[d] {
			return fmt.Errorf("week_days must not contain duplicates")
		}
		seen[d] = true
	}
	return nil
}

func (r WeeklyRecurrence) NextOccurrence(from time.Time) *time.Time {
	from = truncateToDate(from)
	set := make(map[int]bool, len(r.WeekDays))
	for _, d := range r.WeekDays {
		set[d] = true
	}
	for i := 1; i <= 7; i++ {
		candidate := from.AddDate(0, 0, i)
		wd := isoWeekday(candidate)
		if set[wd] {
			return &candidate
		}
	}
	return nil
}

type SpecificDatesRecurrence struct {
	Dates []string `json:"dates"`
}

func (r SpecificDatesRecurrence) RecurrenceType() string { return "specific_dates" }

func (r SpecificDatesRecurrence) Valid() error {
	if len(r.Dates) == 0 {
		return fmt.Errorf("dates must not be empty")
	}
	for _, d := range r.Dates {
		if _, err := time.Parse("2006-01-02", d); err != nil {
			return fmt.Errorf("invalid date %q: must be YYYY-MM-DD", d)
		}
	}
	return nil
}

func (r SpecificDatesRecurrence) NextOccurrence(from time.Time) *time.Time {
	from = truncateToDate(from)
	var earliest *time.Time
	for _, d := range r.Dates {
		t, _ := time.Parse("2006-01-02", d)
		t = truncateToDate(t)
		if t.After(from) {
			if earliest == nil || t.Before(*earliest) {
				copy := t
				earliest = &copy
			}
		}
	}
	return earliest
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}
