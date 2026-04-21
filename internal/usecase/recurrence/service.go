package recurrence

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	"example.com/taskservice/internal/shared/dateutil"
)

const defaultMaterializationHorizonDays = 30

type Service struct {
	repo          Repository
	now           func() time.Time
	horizonInDays int
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:          repo,
		now:           func() time.Time { return time.Now().UTC() },
		horizonInDays: defaultMaterializationHorizonDays,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*recurrencedomain.Rule, error) {
	normalized, err := validateInput(input, dateutil.Today(s.now))
	if err != nil {
		return nil, err
	}

	model := &recurrencedomain.Rule{
		Title:         normalized.Title,
		Description:   normalized.Description,
		Active:        normalized.Active,
		ScheduleType:  normalized.ScheduleType,
		StartsOn:      normalized.StartsOn,
		EndsOn:        normalized.EndsOn,
		EveryNDays:    normalized.EveryNDays,
		DayOfMonth:    normalized.DayOfMonth,
		DayParity:     normalized.DayParity,
		SpecificDates: normalized.SpecificDates,
	}
	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	if err := s.syncRule(ctx, created); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, created.ID)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*recurrencedomain.Rule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateInput(input, dateutil.Today(s.now))
	if err != nil {
		return nil, err
	}

	model := &recurrencedomain.Rule{
		ID:            id,
		Title:         normalized.Title,
		Description:   normalized.Description,
		Active:        normalized.Active,
		ScheduleType:  normalized.ScheduleType,
		StartsOn:      normalized.StartsOn,
		EndsOn:        normalized.EndsOn,
		EveryNDays:    normalized.EveryNDays,
		DayOfMonth:    normalized.DayOfMonth,
		DayParity:     normalized.DayParity,
		SpecificDates: normalized.SpecificDates,
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	if err := s.syncRule(ctx, updated); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, updated.ID)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id, dateutil.Today(s.now))
}

func (s *Service) List(ctx context.Context) ([]recurrencedomain.Rule, error) {
	return s.repo.List(ctx)
}

func (s *Service) MaterializeActive(ctx context.Context) error {
	rules, err := s.repo.ListActive(ctx)
	if err != nil {
		return err
	}

	for i := range rules {
		if err := s.syncRule(ctx, &rules[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) syncRule(ctx context.Context, rule *recurrencedomain.Rule) error {
	fromDate := dateutil.Today(s.now)
	toDate := dateutil.AddDays(fromDate, s.horizonInDays)
	occurrences := rule.OccurrencesBetween(fromDate, toDate)
	return s.repo.SyncGeneratedTasks(ctx, rule, fromDate, occurrences)
}

type normalizedInput struct {
	Title         string
	Description   string
	Active        bool
	ScheduleType  recurrencedomain.ScheduleType
	StartsOn      time.Time
	EndsOn        *time.Time
	EveryNDays    *int
	DayOfMonth    *int
	DayParity     *recurrencedomain.DayParity
	SpecificDates []time.Time
}

func validateInput(input CreateInput, today time.Time) (normalizedInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return normalizedInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.ScheduleType.Valid() {
		return normalizedInput{}, fmt.Errorf("%w: invalid schedule_type", ErrInvalidInput)
	}

	active := true
	if input.Active != nil {
		active = *input.Active
	}

	normalized := normalizedInput{
		Title:        input.Title,
		Description:  input.Description,
		Active:       active,
		ScheduleType: input.ScheduleType,
	}

	switch input.ScheduleType {
	case recurrencedomain.ScheduleTypeEveryNDays:
		if input.StartsOn == nil {
			return normalizedInput{}, fmt.Errorf("%w: starts_on is required for every_n_days", ErrInvalidInput)
		}
		if input.EveryNDays == nil || *input.EveryNDays <= 0 {
			return normalizedInput{}, fmt.Errorf("%w: every_n_days must be positive", ErrInvalidInput)
		}
		normalized.StartsOn = dateutil.Normalize(*input.StartsOn)
		normalized.EveryNDays = input.EveryNDays
		normalized.EndsOn = normalizeOptionalDate(input.EndsOn)
	case recurrencedomain.ScheduleTypeMonthlyDay:
		if input.StartsOn == nil {
			return normalizedInput{}, fmt.Errorf("%w: starts_on is required for monthly_day", ErrInvalidInput)
		}
		if input.DayOfMonth == nil || *input.DayOfMonth < 1 || *input.DayOfMonth > 30 {
			return normalizedInput{}, fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}
		normalized.StartsOn = dateutil.Normalize(*input.StartsOn)
		normalized.DayOfMonth = input.DayOfMonth
		normalized.EndsOn = normalizeOptionalDate(input.EndsOn)
	case recurrencedomain.ScheduleTypeMonthParity:
		if input.StartsOn == nil {
			return normalizedInput{}, fmt.Errorf("%w: starts_on is required for month_day_parity", ErrInvalidInput)
		}
		if input.DayParity == nil || !input.DayParity.Valid() {
			return normalizedInput{}, fmt.Errorf("%w: day_parity must be odd or even", ErrInvalidInput)
		}
		normalized.StartsOn = dateutil.Normalize(*input.StartsOn)
		normalized.DayParity = input.DayParity
		normalized.EndsOn = normalizeOptionalDate(input.EndsOn)
	case recurrencedomain.ScheduleTypeSpecificDates:
		if len(input.SpecificDates) == 0 {
			return normalizedInput{}, fmt.Errorf("%w: dates must not be empty", ErrInvalidInput)
		}

		dedup := make(map[string]time.Time, len(input.SpecificDates))
		for _, rawDate := range input.SpecificDates {
			normalizedDate := dateutil.Normalize(rawDate)
			dedup[dateutil.Format(normalizedDate)] = normalizedDate
		}

		normalizedDates := make([]time.Time, 0, len(dedup))
		for _, value := range dedup {
			normalizedDates = append(normalizedDates, value)
		}

		sort.Slice(normalizedDates, func(i, j int) bool {
			return normalizedDates[i].Before(normalizedDates[j])
		})

		futureOrTodayExists := false
		for _, value := range normalizedDates {
			if !value.Before(today) {
				futureOrTodayExists = true
				break
			}
		}
		if !futureOrTodayExists {
			return normalizedInput{}, fmt.Errorf("%w: dates must contain today or a future date", ErrInvalidInput)
		}

		normalized.SpecificDates = normalizedDates
		normalized.StartsOn = normalizedDates[0]
		last := normalizedDates[len(normalizedDates)-1]
		normalized.EndsOn = &last
	default:
		return normalizedInput{}, fmt.Errorf("%w: unsupported schedule type", ErrInvalidInput)
	}

	if normalized.EndsOn != nil && normalized.EndsOn.Before(normalized.StartsOn) {
		return normalizedInput{}, fmt.Errorf("%w: ends_on must be on or after starts_on", ErrInvalidInput)
	}

	if normalized.ScheduleType != recurrencedomain.ScheduleTypeSpecificDates {
		if normalized.EndsOn != nil && normalized.EndsOn.Before(today) {
			return normalizedInput{}, fmt.Errorf("%w: ends_on must not be in the past", ErrInvalidInput)
		}
	}

	return normalized, nil
}

func normalizeOptionalDate(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	normalized := dateutil.Normalize(*value)
	return &normalized
}
