package task

type DailyIntervalSpec struct{}

func (s DailyIntervalSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Type != RecurrenceTypeDaily {
		return nil
	}
	if cfg.Interval == nil || *cfg.Interval < 1 {
		return ErrInvalidInterval
	}
	return nil
}

type MonthlyDayInRangeSpec struct{}

func (s MonthlyDayInRangeSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Type != RecurrenceTypeMonthly {
		return nil
	}
	if cfg.DayOfMonth == nil || *cfg.DayOfMonth < 1 || *cfg.DayOfMonth > 31 {
		return ErrInvalidDayOfMonth
	}
	return nil
}

type MonthlyOverflowValidSpec struct{}

func (s MonthlyOverflowValidSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Type != RecurrenceTypeMonthly {
		return nil
	}
	if cfg.OverflowStrategy != "" {
		switch cfg.OverflowStrategy {
		case OverflowSkip, OverflowLastDay, OverflowNextMonth:
			return nil
		default:
			return ErrInvalidOverflowStrategy
		}
	}
	return nil
}

type DatesNotEmptySpec struct{}

func (s DatesNotEmptySpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Type != RecurrenceTypeSpecificDates {
		return nil
	}
	if len(cfg.SpecificDates) == 0 {
		return ErrEmptyDates
	}
	return nil
}

type DatesUniqueSpec struct{}

func (s DatesUniqueSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Type != RecurrenceTypeSpecificDates {
		return nil
	}
	seen := make(map[string]struct{}, len(cfg.SpecificDates))
	for _, d := range cfg.SpecificDates {
		if d.IsZero() {
			return ErrInvalidDateInList
		}
		key := d.UTC().Format("2006-01-02")
		if _, exists := seen[key]; exists {
			return ErrDuplicateDate
		}
		seen[key] = struct{}{}
	}
	return nil
}

type ParityModeValidSpec struct{}

func (s ParityModeValidSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Type != RecurrenceTypeParity {
		return nil
	}
	if cfg.ParityMode == nil {
		return ErrInvalidParityMode
	}
	if *cfg.ParityMode != ParityEven && *cfg.ParityMode != ParityOdd {
		return ErrInvalidParityMode
	}
	return nil
}