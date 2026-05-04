package task

import "time"

type RecurrenceType string

const (
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeParity        RecurrenceType = "parity"
)

type ParityMode string

const (
	ParityEven ParityMode = "even"
	ParityOdd  ParityMode = "odd"
)

type OverflowStrategy string

const (
	OverflowSkip      OverflowStrategy = "skip"     
	OverflowLastDay   OverflowStrategy = "last_day"  
	OverflowNextMonth OverflowStrategy = "next_month" 
)

type RecurrenceConfig struct {
	Type RecurrenceType `json:"type"`
	Interval *int `json:"interval,omitempty"`
	DayOfMonth *int `json:"day_of_month,omitempty"`
	SpecificDates []time.Time `json:"specific_dates,omitempty"`
	ParityMode *ParityMode `json:"parity_mode,omitempty"`
	StartDate time.Time `json:"start_date"`
	EndDate *time.Time `json:"end_date,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	OverflowStrategy OverflowStrategy `json:"overflow_strategy,omitempty"`
}

func (c RecurrenceConfig) CalculateEffectiveStart(now time.Time) time.Time {
	if !c.StartDate.Before(now) {
		return c.StartDate
	}

	loc := c.resolveLocation()

	switch c.Type {
	case RecurrenceTypeDaily:
		interval := 1
		if c.Interval != nil {
			interval = *c.Interval
		}

		startDay := c.StartDate.In(loc).Truncate(24 * time.Hour)
		nowDay := now.In(loc).Truncate(24 * time.Hour)
		daysPassed := int(nowDay.Sub(startDay)/(24*time.Hour)) + 1
		steps := (daysPassed + interval - 1) / interval 

		return startDay.AddDate(0, 0, steps*interval)

	case RecurrenceTypeMonthly:
		day := *c.DayOfMonth
		y, m, _ := now.In(loc).Date()
		check := time.Date(y, m, 1, c.StartDate.Hour(), c.StartDate.Minute(), 0, 0, loc)

		for {
			candidate := time.Date(check.Year(), check.Month(), day, check.Hour(), check.Minute(), 0, 0, loc)
			if candidate.Month() == check.Month() && !candidate.Before(now) {
				return candidate
			}
			check = check.AddDate(0, 1, 0)
		}

	case RecurrenceTypeSpecificDates:
		for _, d := range c.SpecificDates {
			if !d.Before(now) {
				return d.In(loc)
			}
		}
		return now.In(loc).Truncate(24 * time.Hour).AddDate(0, 0, 1)

	case RecurrenceTypeParity:
		mode := *c.ParityMode
		current := now.In(loc).Truncate(24 * time.Hour)
		if current.Before(now) {
			current = current.Add(24 * time.Hour)
		}

		for {
			isEven := current.Day()%2 == 0
			if (mode == ParityEven && isEven) || (mode == ParityOdd && !isEven) {
				return current
			}
			current = current.Add(24 * time.Hour)
		}

	default:
		return now.Truncate(24 * time.Hour)
	}
}