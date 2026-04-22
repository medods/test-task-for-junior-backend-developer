package periodicity

import (
	"fmt"
	"time"
)

// PeriodicityType определяет тип стратегии периодичности
type PeriodicityType string

const (
	PeriodicityTypeRRule PeriodicityType = "rrule"
	PeriodicityTypeCron  PeriodicityType = "cron"
	PeriodicityTypeFixed PeriodicityType = "fixed"
)

// Params содержит параметры для создания стратегии
type Params struct {
	Type    PeriodicityType
	RRule   string
	ExDates []string
	RDates  []string
	Cron    string
	Fixed   time.Duration
	Dtstart time.Time // <-- новое поле
}

// Factory создает нужную стратегию по типу периодичности
func Factory(p *Params) (Strategy, error) {
	if p == nil {
		return nil, fmt.Errorf("params is nil")
	}

	switch p.Type {
	case PeriodicityTypeRRule:
		return NewRRuleStrategy(p.RRule, p.ExDates, p.RDates, p.Dtstart)
	case PeriodicityTypeCron:
		return nil, fmt.Errorf("cron strategy not implemented yet")
	case PeriodicityTypeFixed:
		return nil, fmt.Errorf("fixed strategy not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported periodicity type: %s", p.Type)
	}
}