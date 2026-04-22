package periodicity

import (
	//"fmt"
	"time"

	"github.com/teambition/rrule-go"
)

// RRuleStrategy реализует Strategy для rrule-правил
// Использует стандарт RFC 5545
type RRuleStrategy struct {
	RRule   string
	ExDates []string
	RDates  []string
	dtstart time.Time
}

// NewRRuleStrategy создаёт стратегию с фиксированным DTSTART.
func NewRRuleStrategy(rruleStr string, exDates, rDates []string, dtstart time.Time) (*RRuleStrategy, error) {
	// Валидация правила с добавленным DTSTART
	fullRule := rruleStr + ";DTSTART=" + dtstart.UTC().Format("20060102T150405Z")
	if _, err := rrule.StrToRRule(fullRule); err != nil {
		return nil, err
	}
	return &RRuleStrategy{
		RRule:   rruleStr,
		ExDates: exDates,
		RDates:  rDates,
		dtstart: dtstart,
	}, nil
}

// fullRuleString возвращает полное RRULE с DTSTART.
func (s *RRuleStrategy) fullRuleString() string {
	return s.RRule + ";DTSTART=" + s.dtstart.UTC().Format("20060102T150405Z")
}

func (s *RRuleStrategy) Occurrences(from, to time.Time) ([]time.Time, error) {
	rule, err := rrule.StrToRRule(s.fullRuleString())
	if err != nil {
		return nil, err
	}
	dates := rule.Between(from, to, true)

	ex := make(map[string]struct{})
	for _, d := range s.ExDates {
		ex[d] = struct{}{}
	}
	var result []time.Time
	for _, d := range dates {
		if _, skip := ex[d.Format("2006-01-02")]; !skip {
			result = append(result, d)
		}
	}
	for _, rd := range s.RDates {
		t, err := time.Parse("2006-01-02", rd)
		if err == nil && t.After(from) && t.Before(to) {
			result = append(result, t)
		}
	}
	return result, nil
}

func (s *RRuleStrategy) Type() string {
	return "rrule"
}

func (s *RRuleStrategy) NextAfter(after time.Time) (time.Time, error) {
	rule, err := rrule.StrToRRule(s.fullRuleString())
	if err != nil {
		return time.Time{}, err
	}

	// Получаем следующее вхождение после after
	next := rule.After(after, false)

	// Если next нулевое, значит вхождений больше нет по RRULE
	if next.IsZero() {
		// Проверяем RDates вручную
		var candidate time.Time
		for _, rd := range s.RDates {
			t, err := time.Parse("2006-01-02", rd)
			if err == nil && t.After(after) {
				if candidate.IsZero() || t.Before(candidate) {
					candidate = t
				}
			}
		}
		if candidate.IsZero() {
			return time.Time{}, nil // Нет больше дат
		}
		return candidate, nil
	}

	// Проверяем ExDates
	exMap := make(map[string]struct{}, len(s.ExDates))
	for _, d := range s.ExDates {
		exMap[d] = struct{}{}
	}
	if _, excluded := exMap[next.Format("2006-01-02")]; excluded {
		// Рекурсивно ищем следующее после исключённой даты
		return s.NextAfter(next)
	}

	// Проверяем, есть ли более раннее RDate
	for _, rd := range s.RDates {
		t, err := time.Parse("2006-01-02", rd)
		if err == nil && t.After(after) && t.Before(next) && !t.Equal(next) {
			// Проверяем, не исключена ли эта RDate
			if _, ex := exMap[t.Format("2006-01-02")]; !ex {
				next = t
			}
		}
	}

	return next, nil
}