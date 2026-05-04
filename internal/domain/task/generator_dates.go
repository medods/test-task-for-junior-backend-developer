package task

import "time"

type DatesGenerator struct{}

func (g *DatesGenerator) Generate(cfg RecurrenceConfig) ([]time.Time, error) {
	if len(cfg.SpecificDates) == 0 {
		return nil, ErrEmptyDates
	}

	var dates []time.Time
	for _, d := range cfg.SpecificDates {
		normalized := d.UTC().Truncate(24 * time.Hour)
		if !normalized.Before(cfg.StartDate) && (cfg.EndDate == nil || !normalized.After(*cfg.EndDate)) {
			dates = append(dates, normalized)
		}
	}

	return dates, nil
}