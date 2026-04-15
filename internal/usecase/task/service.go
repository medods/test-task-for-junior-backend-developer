// Package task
package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/taskservice/internal/domain/recurrence"
	task "example.com/taskservice/internal/domain/task"
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

func (s *Service) Create(ctx context.Context, input CreateInput) (*task.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &task.Task{
		Title:               normalized.Title,
		Description:         normalized.Description,
		Status:              normalized.Status,
		RecurrenceType:      normalized.RecurrenceType,
		RecurrencePayload:   normalized.RecurrencePayload,
		RecurrenceStartDate: normalized.RecurrenceStartDate,
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

func (s *Service) GetByID(ctx context.Context, id int64) (*task.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*task.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &task.Task{
		ID:                  id,
		Title:               normalized.Title,
		Description:         normalized.Description,
		Status:              normalized.Status,
		RecurrenceType:      normalized.RecurrenceType,
		RecurrencePayload:   normalized.RecurrencePayload,
		RecurrenceStartDate: normalized.RecurrenceStartDate,
		UpdatedAt:           s.now(),
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

func (s *Service) List(ctx context.Context) ([]task.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByDate(ctx context.Context, date time.Time) ([]task.Task, error) {
	return s.repo.ListByDate(ctx, date)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = task.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.RecurrenceType != recurrence.NoneRecurrence && input.RecurrenceStartDate == nil {
		return CreateInput{}, fmt.Errorf("%w: invalid_recurrence type or recurrence_start_date", ErrInvalidInput)
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

	if input.RecurrenceType != recurrence.NoneRecurrence && input.RecurrenceStartDate == nil {
		return UpdateInput{}, fmt.Errorf("%w: invalid_recurrence type or recurrence_start_date", ErrInvalidInput)
	}

	return input, nil
}
