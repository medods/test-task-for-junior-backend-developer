package task

import "time"

type MonthlyGenerator struct{}

func (g *MonthlyGenerator) Generate(cfg RecurrenceConfig) ([]time.Time, error) {
	if cfg.DayOfMonth == nil || *cfg.DayOfMonth < 1 || *cfg.DayOfMonth > 31 {
		return nil, ErrInvalidDayOfMonth
	}

	var dates []time.Time
	current := cfg.StartDate.UTC().Truncate(24 * time.Hour)
	limit := calculateGenerationLimit(cfg.EndDate)
	day := *cfg.DayOfMonth
	strategy := cfg.OverflowStrategy
	if strategy == "" {
		strategy = OverflowSkip
	}

	for len(dates) < limit && (cfg.EndDate == nil || !current.After(*cfg.EndDate)) {
		y, m, _ := current.Date()
		h, min, sec := current.Clock()

		candidate := time.Date(y, m, day, h, min, sec, 0, time.UTC)

		if candidate.Day() != day {
			switch strategy {
			case OverflowSkip:
			case OverflowLastDay:
				lastDay := time.Date(y, m+1, 0, h, min, sec, 0, time.UTC).Day()
				candidate = time.Date(y, m, lastDay, h, min, sec, 0, time.UTC)
				dates = append(dates, candidate)
			case OverflowNextMonth:
				candidate = time.Date(y, m+1, 1, h, min, sec, 0, time.UTC)
				if cfg.EndDate == nil || !candidate.After(*cfg.EndDate) {
					dates = append(dates, candidate)
				}
			}
		} else {
			dates = append(dates, candidate)
		}

		if m == time.December {
			current = time.Date(y+1, time.January, 1, 0, 0, 0, 0, time.UTC)
		} else {
			current = time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return dates, nil
}