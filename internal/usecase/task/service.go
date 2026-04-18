package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	if normalized.Recurrence == nil {
		model := &taskdomain.Task{
			Title:        normalized.Title,
			Description:  normalized.Description,
			Status:       normalized.Status,
			ScheduledFor: normalized.ScheduledFor,
		}

		now := s.now()
		model.CreatedAt = now
		model.UpdatedAt = now

		created, err := s.repo.Create(ctx, model)
		if err != nil {
			return nil, err
		}

		return created, nil
	}

	if normalized.Recurrence.Type == taskdomain.RecurrenceSpecificDates {
		if len(normalized.Recurrence.SpecificDates) == 0 {
			return nil, fmt.Errorf("%w: specific_dates is empty", ErrInvalidInput)
		}

		var firstCreated *taskdomain.Task

		for _, date := range normalized.Recurrence.SpecificDates {
			currentDate := date

			model := &taskdomain.Task{
				Title:        normalized.Title,
				Description:  normalized.Description,
				Status:       normalized.Status,
				ScheduledFor: &currentDate,
			}

			now := s.now()
			model.CreatedAt = now
			model.UpdatedAt = now

			created, err := s.repo.Create(ctx, model)
			if err != nil {
				return nil, err
			}

			if firstCreated == nil {
				firstCreated = created
			}
		}

		return firstCreated, nil
	}

	if normalized.Recurrence.Type == taskdomain.RecurrenceDaily {
		if normalized.Recurrence.EndDate == nil {
			return nil, fmt.Errorf("%w: end_date is required for daily recurrence", ErrInvalidInput)
		}

		if normalized.Recurrence.EveryNDays == nil {
			return nil, fmt.Errorf("%w: every_n_days is required for daily recurrence", ErrInvalidInput)
		}

		if *normalized.Recurrence.EveryNDays <= 0 {
			return nil, fmt.Errorf("%w: every_n_days must be greater than zero", ErrInvalidInput)
		}

		var firstCreated *taskdomain.Task

		for date := normalized.Recurrence.StartDate; !date.After(*normalized.Recurrence.EndDate); date = date.AddDate(0, 0, *normalized.Recurrence.EveryNDays) {
			currentDate := date

			model := &taskdomain.Task{
				Title:        normalized.Title,
				Description:  normalized.Description,
				Status:       normalized.Status,
				ScheduledFor: &currentDate,
			}

			now := s.now()
			model.CreatedAt = now
			model.UpdatedAt = now

			created, err := s.repo.Create(ctx, model)
			if err != nil {
				return nil, err
			}

			if firstCreated == nil {
				firstCreated = created
			}
		}

		return firstCreated, nil
	}

	if normalized.Recurrence.Type == taskdomain.RecurrenceOddDays ||
		normalized.Recurrence.Type == taskdomain.RecurrenceEvenDays {

		if normalized.Recurrence.EndDate == nil {
			return nil, fmt.Errorf("%w: end_date is required for odd/even recurrence", ErrInvalidInput)
		}

		var firstCreated *taskdomain.Task

		for date := normalized.Recurrence.StartDate; !date.After(*normalized.Recurrence.EndDate); date = date.AddDate(0, 0, 1) {
			day := date.Day()

			if normalized.Recurrence.Type == taskdomain.RecurrenceOddDays && day%2 == 0 {
				continue
			}

			if normalized.Recurrence.Type == taskdomain.RecurrenceEvenDays && day%2 != 0 {
				continue
			}

			currentDate := date

			model := &taskdomain.Task{
				Title:        normalized.Title,
				Description:  normalized.Description,
				Status:       normalized.Status,
				ScheduledFor: &currentDate,
			}

			now := s.now()
			model.CreatedAt = now
			model.UpdatedAt = now

			created, err := s.repo.Create(ctx, model)
			if err != nil {
				return nil, err
			}

			if firstCreated == nil {
				firstCreated = created
			}
		}

		return firstCreated, nil
	}

		if normalized.Recurrence.Type == taskdomain.RecurrenceMonthly {
		if normalized.Recurrence.EndDate == nil {
			return nil, fmt.Errorf("%w: end_date is required for monthly recurrence", ErrInvalidInput)
		}

		if normalized.Recurrence.DayOfMonth == nil {
			return nil, fmt.Errorf("%w: day_of_month is required for monthly recurrence", ErrInvalidInput)
		}

		if *normalized.Recurrence.DayOfMonth < 1 || *normalized.Recurrence.DayOfMonth > 30 {
			return nil, fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}

		var firstCreated *taskdomain.Task

		start := normalized.Recurrence.StartDate
		end := *normalized.Recurrence.EndDate

		current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)

		for !current.After(end) {
			candidate := time.Date(
				current.Year(),
				current.Month(),
				*normalized.Recurrence.DayOfMonth,
				0, 0, 0, 0,
				time.UTC,
			)

			if !candidate.Before(start) && !candidate.After(end) {
				currentDate := candidate

				model := &taskdomain.Task{
					Title:        normalized.Title,
					Description:  normalized.Description,
					Status:       normalized.Status,
					ScheduledFor: &currentDate,
				}

				now := s.now()
				model.CreatedAt = now
				model.UpdatedAt = now

				created, err := s.repo.Create(ctx, model)
				if err != nil {
					return nil, err
				}

				if firstCreated == nil {
					firstCreated = created
				}
			}

			current = current.AddDate(0, 1, 0)
		}

		if firstCreated == nil {
			return nil, fmt.Errorf("%w: no dates generated for monthly recurrence", ErrInvalidInput)
		}

		return firstCreated, nil
	}

	return nil, fmt.Errorf("recurrence type is not implemented yet")

}


func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
