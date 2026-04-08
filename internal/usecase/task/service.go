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

	model := &taskdomain.Task{
		Title:            normalized.Title,
		Description:      normalized.Description,
		Status:           normalized.Status,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceEveryN: normalized.RecurrenceEveryN,
		MonthlyDay:       normalized.MonthlyDay,
		SpecificDates:    normalized.SpecificDates,
		EvenOdd:          normalized.EvenOdd,
	}

	// ВАЖНО: Валидируем периодичность перед сохранением
	if err := model.ValidateRecurrence(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	// Устанавливаем статус по умолчанию, если не указан
	if model.Status == "" {
		model.Status = taskdomain.StatusNew
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
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

	// Сначала получаем существующую задачу
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	// Обновляем только те поля, которые переданы
	if normalized.Title != nil {
		existing.Title = *normalized.Title
	}
	if normalized.Description != nil {
		existing.Description = *normalized.Description
	}
	if normalized.Status != nil && normalized.Status.Valid() {
		existing.Status = *normalized.Status
	}

	// Обновляем поля периодичности, если они переданы
	if normalized.RecurrenceType != nil {
		existing.RecurrenceType = *normalized.RecurrenceType
		existing.RecurrenceEveryN = normalized.RecurrenceEveryN
		existing.MonthlyDay = normalized.MonthlyDay
		existing.SpecificDates = normalized.SpecificDates
		existing.EvenOdd = normalized.EvenOdd

		// Валидируем обновленную периодичность
		if err := existing.ValidateRecurrence(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	existing.UpdatedAt = s.now()

	updated, err := s.repo.Update(ctx, existing)
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
	// Обработка Title (указатель)
	if input.Title != nil {
		trimmed := strings.TrimSpace(*input.Title)
		if trimmed == "" {
			return UpdateInput{}, fmt.Errorf("%w: title cannot be empty", ErrInvalidInput)
		}
		input.Title = &trimmed
	}

	// Обработка Description (указатель)
	if input.Description != nil {
		trimmed := strings.TrimSpace(*input.Description)
		input.Description = &trimmed
	}

	// Валидация статуса (указатель)
	if input.Status != nil && !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
