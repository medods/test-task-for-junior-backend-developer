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
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
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

func (s *Service) GetRecurrence(ctx context.Context, taskID int64) (int64, taskdomain.Recurrence, time.Time, error) {
	if taskID <= 0 {
		return 0, nil, time.Time{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return 0, nil, time.Time{}, err
	}

	if task.RecurrenceID == nil {
		return 0, nil, time.Time{}, taskdomain.ErrRecurrenceNotFound
	}

	rec, nextAt, err := s.repo.GetRecurrence(ctx, *task.RecurrenceID)
	if err != nil {
		return 0, nil, time.Time{}, err
	}

	return *task.RecurrenceID, rec, nextAt, nil
}

func (s *Service) SetRecurrence(ctx context.Context, taskID int64, rec taskdomain.Recurrence) error {
	if taskID <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if err := rec.Valid(); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	nextAt := rec.NextOccurrence(s.now())
	if nextAt == nil {
		return fmt.Errorf("%w: recurrence has no future occurrences", ErrInvalidInput)
	}

	if task.RecurrenceID != nil {
		if err := s.repo.DeleteRecurrence(ctx, *task.RecurrenceID); err != nil {
			return err
		}
	}

	recurrenceID, err := s.repo.CreateRecurrence(ctx, rec, *nextAt)
	if err != nil {
		return err
	}

	return s.repo.SetTaskRecurrence(ctx, taskID, &recurrenceID)
}

func (s *Service) DeleteRecurrence(ctx context.Context, taskID int64) error {
	if taskID <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if task.RecurrenceID == nil {
		return taskdomain.ErrRecurrenceNotFound
	}

	return s.repo.DeleteRecurrence(ctx, *task.RecurrenceID)
}

