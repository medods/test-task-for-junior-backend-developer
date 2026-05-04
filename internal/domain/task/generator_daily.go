package task

import "time"

type DailyGenerator struct{}

func (g *DailyGenerator) Generate(cfg RecurrenceConfig) ([]time.Time, error) {
	if cfg.Interval == nil || *cfg.Interval < 1 {
		return nil, ErrInvalidInterval
	}

	var dates []time.Time
	current := cfg.StartDate.UTC().Truncate(24 * time.Hour)
	limit := calculateGenerationLimit(cfg.EndDate)

	for len(dates) < limit && (cfg.EndDate == nil || !current.After(*cfg.EndDate)) {
		dates = append(dates, current)
		current = current.AddDate(0, 0, *cfg.Interval)
	}

	return dates, nil
}