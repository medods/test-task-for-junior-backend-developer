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

	if normalized.Recurrence != nil {
		if err := normalized.Recurrence.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	if !normalized.ScheduledAt.IsZero() && normalized.ScheduledAt.Before(s.now()) {
		return nil, fmt.Errorf("%w: scheduled_at cannot be in the past", ErrInvalidInput)
	}

	model := &taskdomain.Task{
		Title:               normalized.Title,
		Description:         normalized.Description,
		Status:              normalized.Status,
		CreatedAt:           s.now(),
		UpdatedAt:           s.now(),
		ScheduledAt:         normalized.ScheduledAt,
		Recurrence:          normalized.Recurrence,
		IsRecurringTemplate: normalized.Recurrence != nil,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	if created.IsRecurringTemplate && created.Recurrence != nil {
		dates, err := taskdomain.GenerateRecurringDates(created.Recurrence, created.ScheduledAt, s.now())
		if err != nil {
			// Логируем ошибку, но не отменяем создание шаблона
			return created, nil
		}

		for _, date := range dates {
			instance := &taskdomain.Task{
				Title:            created.Title,
				Description:      created.Description,
				Status:           taskdomain.StatusNew,
				ScheduledAt:      date,
				ParentTemplateID: &created.ID,
				CreatedAt:        s.now(),
				UpdatedAt:        s.now(),
			}
			_, err := s.repo.Create(ctx, instance)
			if err != nil {
				// Логируем ошибку, но продолжаем
				continue
			}
		}
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) generateInstances(ctx context.Context, template *taskdomain.Task, now time.Time) error {
	if template.Recurrence == nil {
		return nil
	}

	// Генерируем даты для экземпляров
	dates, err := taskdomain.GenerateRecurringDates(template.Recurrence, template.ScheduledAt, now)
	if err != nil {
		return err
	}

	// Для каждой даты создаём задачу-экземпляр
	for _, date := range dates {
		instance := &taskdomain.Task{
			Title:            template.Title,
			Description:      template.Description,
			Status:           taskdomain.StatusNew,
			ScheduledAt:      date,
			ParentTemplateID: &template.ID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if _, err := s.repo.Create(ctx, instance); err != nil {
			// Логируем ошибку, но продолжаем создавать остальные экземпляры
			continue
		}
	}
	return nil
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

	if normalized.Recurrence != nil {
		if err := normalized.Recurrence.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	if !normalized.ScheduledAt.IsZero() && normalized.ScheduledAt.Before(s.now()) {
		return nil, fmt.Errorf("%w: scheduled_at cannot be in the past", ErrInvalidInput)
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}
	// ScheduledAt: если передано новое - используем его, иначе оставляем старое
	if !normalized.ScheduledAt.IsZero() {
		model.ScheduledAt = normalized.ScheduledAt
	} else {
		model.ScheduledAt = existing.ScheduledAt
	}
	// Recurrence: если передано новое - используем его, иначе оставляем старое
	if normalized.Recurrence != nil {
		model.Recurrence = normalized.Recurrence
	} else {
		model.Recurrence = existing.Recurrence
	}

	model.IsRecurringTemplate = model.Recurrence != nil

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	if updated.IsRecurringTemplate && updated.Recurrence != nil {
		// Удаляем все старые экземпляры, у которых scheduled_at >= текущего времени
		instances, _ := s.repo.GetInstancesByTemplateID(ctx, updated.ID)
		now := s.now()
		for _, inst := range instances {
			if !inst.ScheduledAt.Before(now) {
				_ = s.repo.Delete(ctx, inst.ID)
			}
		}
		// Создаём новые экземпляры
		_ = s.generateInstances(ctx, updated, now)
	} else if !updated.IsRecurringTemplate && existing.Recurrence != nil {
		// Если шаблон перестал быть шаблоном (recurrence удалён) — удаляем все его экземпляры
		instances, _ := s.repo.GetInstancesByTemplateID(ctx, updated.ID)
		for _, inst := range instances {
			_ = s.repo.Delete(ctx, inst.ID)
		}
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
