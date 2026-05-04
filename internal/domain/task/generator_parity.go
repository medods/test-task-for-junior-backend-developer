package task

import "time"

type ParityGenerator struct{}

func (g *ParityGenerator) Generate(cfg RecurrenceConfig) ([]time.Time, error) {
	if cfg.ParityMode == nil {
		return nil, ErrInvalidParityMode
	}
	mode := *cfg.ParityMode

	var dates []time.Time
	current := cfg.StartDate.UTC().Truncate(24 * time.Hour)
	limit := calculateGenerationLimit(cfg.EndDate)

	for len(dates) < limit && (cfg.EndDate == nil || !current.After(*cfg.EndDate)) {
		isEven := current.Day()%2 == 0
		if (mode == ParityEven && isEven) || (mode == ParityOdd && !isEven) {
			dates = append(dates, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return dates, nil
}