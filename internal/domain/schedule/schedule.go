package schedule

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

type RuleType string

const (
	RuleTypeDailyInterval RuleType = "daily_interval"
	RuleTypeMonthlyDay    RuleType = "monthly_day"
	RuleTypeSpecificDates RuleType = "specific_dates"
	RuleTypeMonthParity   RuleType = "month_parity"
	RuleTypeWeekly        RuleType = "weekly"
)

type MonthParity string

const (
	MonthParityEven MonthParity = "even"
	MonthParityOdd  MonthParity = "odd"
)

type TaskSchedule struct {
	ID          int64
	Title       string
	Description string
	Rule        RecurrenceRule
	StartDate   time.Time
	EndDate     *time.Time
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RecurrenceRule interface {
	Type() RuleType
	Validate() error
	Occurrences(startDate time.Time, endDate *time.Time, from time.Time, to time.Time) []time.Time
}

type DailyIntervalRule struct {
	EveryDays int `json:"every_days"`
}

type MonthlyDayRule struct {
	DayOfMonth int `json:"day_of_month"`
}

type SpecificDatesRule struct {
	Dates []time.Time
}

type MonthParityRule struct {
	Parity MonthParity `json:"parity"`
}

type specificDatesPayload struct {
	Dates []string `json:"dates"`
}

func NormalizeDate(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func DateFromLocation(now time.Time, location *time.Location) time.Time {
	local := now.In(location)
	return NormalizeDate(time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC))
}

func ParseDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}

	return NormalizeDate(parsed), nil
}

func FormatDate(value time.Time) string {
	return NormalizeDate(value).Format("2006-01-02")
}

func (s *TaskSchedule) Normalize() {
	s.Title = strings.TrimSpace(s.Title)
	s.Description = strings.TrimSpace(s.Description)
	s.StartDate = NormalizeDate(s.StartDate)
	if s.EndDate != nil {
		normalized := NormalizeDate(*s.EndDate)
		s.EndDate = &normalized
	}
}

func (s TaskSchedule) Validate() error {
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if s.Rule == nil {
		return fmt.Errorf("recurrence rule is required")
	}

	if err := s.Rule.Validate(); err != nil {
		return err
	}

	if s.StartDate.IsZero() {
		return fmt.Errorf("start_date is required")
	}

	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return fmt.Errorf("end_date must be greater than or equal to start_date")
	}

	return nil
}

func (s TaskSchedule) OccurrenceDates(from time.Time, to time.Time) []time.Time {
	if s.Rule == nil {
		return nil
	}

	return s.Rule.Occurrences(s.StartDate, s.EndDate, NormalizeDate(from), NormalizeDate(to))
}

func (r DailyIntervalRule) Type() RuleType {
	return RuleTypeDailyInterval
}

func (r DailyIntervalRule) Validate() error {
	if r.EveryDays <= 0 {
		return fmt.Errorf("every_days must be greater than 0")
	}

	return nil
}

func (r DailyIntervalRule) Occurrences(startDate time.Time, endDate *time.Time, from time.Time, to time.Time) []time.Time {
	windowStart, windowEnd, ok := boundedWindow(startDate, endDate, from, to)
	if !ok {
		return nil
	}

	first := startDate
	if windowStart.After(startDate) {
		diffDays := int(windowStart.Sub(startDate).Hours() / 24)
		remainder := diffDays % r.EveryDays
		first = windowStart
		if remainder != 0 {
			first = windowStart.AddDate(0, 0, r.EveryDays-remainder)
		}
	}

	result := make([]time.Time, 0)
	for current := first; !current.After(windowEnd); current = current.AddDate(0, 0, r.EveryDays) {
		result = append(result, NormalizeDate(current))
	}

	return result
}

func (r MonthlyDayRule) Type() RuleType {
	return RuleTypeMonthlyDay
}

func (r MonthlyDayRule) Validate() error {
	if r.DayOfMonth < 1 || r.DayOfMonth > 30 {
		return fmt.Errorf("day_of_month must be between 1 and 30")
	}

	return nil
}

func (r MonthlyDayRule) Occurrences(startDate time.Time, endDate *time.Time, from time.Time, to time.Time) []time.Time {
	windowStart, windowEnd, ok := boundedWindow(startDate, endDate, from, to)
	if !ok {
		return nil
	}

	currentMonth := time.Date(windowStart.Year(), windowStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonth := time.Date(windowEnd.Year(), windowEnd.Month(), 1, 0, 0, 0, 0, time.UTC)

	result := make([]time.Time, 0)
	for monthCursor := currentMonth; !monthCursor.After(lastMonth); monthCursor = monthCursor.AddDate(0, 1, 0) {
		lastDay := daysInMonth(monthCursor.Year(), monthCursor.Month())
		if r.DayOfMonth > lastDay {
			continue
		}

		candidate := time.Date(monthCursor.Year(), monthCursor.Month(), r.DayOfMonth, 0, 0, 0, 0, time.UTC)
		if candidate.Before(windowStart) || candidate.After(windowEnd) {
			continue
		}

		result = append(result, candidate)
	}

	return result
}

func (r SpecificDatesRule) Type() RuleType {
	return RuleTypeSpecificDates
}

func (r SpecificDatesRule) Validate() error {
	if len(r.Dates) == 0 {
		return fmt.Errorf("dates must not be empty")
	}

	seen := make(map[string]struct{}, len(r.Dates))
	for _, raw := range r.Dates {
		normalized := NormalizeDate(raw)
		key := FormatDate(normalized)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("dates must not contain duplicates")
		}

		seen[key] = struct{}{}
	}

	return nil
}

func (r SpecificDatesRule) Occurrences(startDate time.Time, endDate *time.Time, from time.Time, to time.Time) []time.Time {
	windowStart, windowEnd, ok := boundedWindow(startDate, endDate, from, to)
	if !ok {
		return nil
	}

	result := make([]time.Time, 0, len(r.Dates))
	for _, raw := range r.Dates {
		candidate := NormalizeDate(raw)
		if candidate.Before(windowStart) || candidate.After(windowEnd) {
			continue
		}

		result = append(result, candidate)
	}

	slices.SortFunc(result, func(left time.Time, right time.Time) int {
		switch {
		case left.Before(right):
			return -1
		case left.After(right):
			return 1
		default:
			return 0
		}
	})

	return result
}

func (r MonthParityRule) Type() RuleType {
	return RuleTypeMonthParity
}

func (r MonthParityRule) Validate() error {
	switch r.Parity {
	case MonthParityEven, MonthParityOdd:
		return nil
	default:
		return fmt.Errorf("parity must be even or odd")
	}
}

func (r MonthParityRule) Occurrences(startDate time.Time, endDate *time.Time, from time.Time, to time.Time) []time.Time {
	windowStart, windowEnd, ok := boundedWindow(startDate, endDate, from, to)
	if !ok {
		return nil
	}

	result := make([]time.Time, 0)
	for current := windowStart; !current.After(windowEnd); current = current.AddDate(0, 0, 1) {
		if r.Parity == MonthParityEven && current.Day()%2 == 0 {
			result = append(result, current)
		}
		if r.Parity == MonthParityOdd && current.Day()%2 != 0 {
			result = append(result, current)
		}
	}

	return result
}

func EncodeRule(rule RecurrenceRule) (RuleType, []byte, error) {
	switch typed := rule.(type) {
	case DailyIntervalRule:
		payload, err := json.Marshal(typed)
		return typed.Type(), payload, err
	case *DailyIntervalRule:
		payload, err := json.Marshal(typed)
		return typed.Type(), payload, err
	case MonthlyDayRule:
		payload, err := json.Marshal(typed)
		return typed.Type(), payload, err
	case *MonthlyDayRule:
		payload, err := json.Marshal(typed)
		return typed.Type(), payload, err
	case SpecificDatesRule:
		payload, err := json.Marshal(specificDatesPayload{Dates: formatDates(typed.Dates)})
		return typed.Type(), payload, err
	case *SpecificDatesRule:
		payload, err := json.Marshal(specificDatesPayload{Dates: formatDates(typed.Dates)})
		return typed.Type(), payload, err
	case MonthParityRule:
		payload, err := json.Marshal(typed)
		return typed.Type(), payload, err
	case *MonthParityRule:
		payload, err := json.Marshal(typed)
		return typed.Type(), payload, err
	default:
		return "", nil, fmt.Errorf("unsupported recurrence rule type")
	}
}

func DecodeRule(ruleType RuleType, payload []byte) (RecurrenceRule, error) {
	switch ruleType {
	case RuleTypeDailyInterval:
		var rule DailyIntervalRule
		if err := json.Unmarshal(payload, &rule); err != nil {
			return nil, err
		}
		return rule, nil
	case RuleTypeMonthlyDay:
		var rule MonthlyDayRule
		if err := json.Unmarshal(payload, &rule); err != nil {
			return nil, err
		}
		return rule, nil
	case RuleTypeSpecificDates:
		var raw specificDatesPayload
		if err := json.Unmarshal(payload, &raw); err != nil {
			return nil, err
		}

		dates := make([]time.Time, 0, len(raw.Dates))
		for _, value := range raw.Dates {
			parsed, err := ParseDate(value)
			if err != nil {
				return nil, err
			}

			dates = append(dates, parsed)
		}

		return SpecificDatesRule{Dates: dates}, nil
	case RuleTypeMonthParity:
		var rule MonthParityRule
		if err := json.Unmarshal(payload, &rule); err != nil {
			return nil, err
		}
		return rule, nil
	case RuleTypeWeekly:
		return nil, fmt.Errorf("weekly recurrence is reserved for future versions")
	default:
		return nil, fmt.Errorf("unsupported recurrence rule type")
	}
}

func boundedWindow(startDate time.Time, endDate *time.Time, from time.Time, to time.Time) (time.Time, time.Time, bool) {
	windowStart := NormalizeDate(from)
	windowEnd := NormalizeDate(to)
	scheduleStart := NormalizeDate(startDate)

	if scheduleStart.After(windowStart) {
		windowStart = scheduleStart
	}

	if endDate != nil {
		scheduleEnd := NormalizeDate(*endDate)
		if scheduleEnd.Before(windowEnd) {
			windowEnd = scheduleEnd
		}
	}

	if windowEnd.Before(windowStart) {
		return time.Time{}, time.Time{}, false
	}

	return windowStart, windowEnd, true
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func formatDates(dates []time.Time) []string {
	formatted := make([]string, 0, len(dates))
	for _, value := range dates {
		formatted = append(formatted, FormatDate(value))
	}

	return formatted
}
