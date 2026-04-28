package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

const instanceGenerationHorizonDays = 30

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

	now := s.now()

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Mode:        normalized.Mode,
		Recurrence:  normalized.Recurrence,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	switch normalized.Mode {
	case taskdomain.ModeOneTime:
		return s.repo.Create(ctx, model)

	case taskdomain.ModeRecurringTemplate:
		scheduledDates := generateScheduledDates(
			normalized.Recurrence,
			now,
			instanceGenerationHorizonDays,
		)

		return s.repo.CreateRecurringTemplate(ctx, model, scheduledDates)

	case taskdomain.ModeRecurringInstance:
		return nil, fmt.Errorf("%w: recurring instances cannot be created directly", ErrInvalidInput)

	default:
		return nil, fmt.Errorf("%w: invalid mode", ErrInvalidInput)
	}
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

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	if normalized.Mode != nil && *normalized.Mode != existing.Mode {
		return nil, fmt.Errorf("%w: mode cannot be changed", ErrInvalidInput)
	}

	now := s.now()

	switch existing.Mode {
	case taskdomain.ModeOneTime:
		if normalized.Recurrence != nil {
			return nil, fmt.Errorf("%w: one-time task cannot have recurrence", ErrInvalidInput)
		}

		model := &taskdomain.Task{
			ID:           existing.ID,
			Title:        normalized.Title,
			Description:  normalized.Description,
			Status:       normalized.Status,
			Mode:         taskdomain.ModeOneTime,
			ScheduledFor: existing.ScheduledFor,
			SourceTaskID: nil,
			Recurrence:   nil,
			CreatedAt:    existing.CreatedAt,
			UpdatedAt:    now,
		}

		return s.repo.Update(ctx, model)

	case taskdomain.ModeRecurringTemplate:
		recurrence := normalized.Recurrence
		if recurrence == nil {
			recurrence = existing.Recurrence
		}

		recurrence, err = validateAndNormalizeRecurrence(recurrence)
		if err != nil {
			return nil, err
		}

		template := &taskdomain.Task{
			ID:           existing.ID,
			Title:        normalized.Title,
			Description:  normalized.Description,
			Status:       normalized.Status,
			Mode:         taskdomain.ModeRecurringTemplate,
			ScheduledFor: nil,
			SourceTaskID: nil,
			Recurrence:   recurrence,
			CreatedAt:    existing.CreatedAt,
			UpdatedAt:    now,
		}

		regenerateFrom := dateOnly(now)
		scheduledDates := generateScheduledDates(
			recurrence,
			regenerateFrom,
			instanceGenerationHorizonDays,
		)

		return s.repo.UpdateRecurringTemplate(
			ctx,
			template,
			&regenerateFrom,
			scheduledDates,
		)

	case taskdomain.ModeRecurringInstance:
		if normalized.Recurrence != nil {
			return nil, fmt.Errorf("%w: recurring instance cannot have recurrence", ErrInvalidInput)
		}

		model := &taskdomain.Task{
			ID:           existing.ID,
			Title:        normalized.Title,
			Description:  normalized.Description,
			Status:       normalized.Status,
			Mode:         taskdomain.ModeRecurringInstance,
			ScheduledFor: existing.ScheduledFor,
			SourceTaskID: existing.SourceTaskID,
			Recurrence:   nil,
			CreatedAt:    existing.CreatedAt,
			UpdatedAt:    now,
		}

		return s.repo.Update(ctx, model)

	default:
		return nil, fmt.Errorf("%w: invalid existing task mode", ErrInvalidInput)
	}
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	switch existing.Mode {
	case taskdomain.ModeOneTime, taskdomain.ModeRecurringInstance:
		return s.repo.Delete(ctx, id)

	case taskdomain.ModeRecurringTemplate:
		return s.repo.DeleteRecurringTemplate(ctx, id)

	default:
		return fmt.Errorf("%w: invalid existing task mode", ErrInvalidInput)
	}
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	if err := s.syncRecurringTemplates(ctx); err != nil {
		return nil, err
	}

	return s.repo.List(ctx)
}

func (s *Service) syncRecurringTemplates(ctx context.Context) error {
	templates, err := s.repo.ListRecurringTemplates(ctx)
	if err != nil {
		return err
	}

	now := s.now()

	for i := range templates {
		template := templates[i]

		if template.Mode != taskdomain.ModeRecurringTemplate {
			continue
		}

		recurrence, err := validateAndNormalizeRecurrence(template.Recurrence)
		if err != nil {
			return err
		}

		template.Recurrence = recurrence

		scheduledDates := generateScheduledDates(
			recurrence,
			now,
			instanceGenerationHorizonDays,
		)

		if len(scheduledDates) == 0 {
			continue
		}

		if _, err := s.repo.CreateRecurringInstances(ctx, &template, scheduledDates); err != nil {
			return err
		}
	}

	return nil
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

	if input.Mode == "" {
		input.Mode = taskdomain.ModeOneTime
	}

	if !input.Mode.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid mode", ErrInvalidInput)
	}

	switch input.Mode {
	case taskdomain.ModeOneTime:
		if input.Recurrence != nil {
			return CreateInput{}, fmt.Errorf("%w: one-time task cannot have recurrence", ErrInvalidInput)
		}

	case taskdomain.ModeRecurringTemplate:
		recurrence, err := validateAndNormalizeRecurrence(input.Recurrence)
		if err != nil {
			return CreateInput{}, err
		}

		input.Recurrence = recurrence

	case taskdomain.ModeRecurringInstance:
		return CreateInput{}, fmt.Errorf("%w: recurring instances cannot be created directly", ErrInvalidInput)
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

	if input.Mode != nil && !(*input.Mode).Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid mode", ErrInvalidInput)
	}

	if input.Recurrence != nil {
		recurrence, err := validateAndNormalizeRecurrence(input.Recurrence)
		if err != nil {
			return UpdateInput{}, err
		}

		input.Recurrence = recurrence
	}

	return input, nil
}
