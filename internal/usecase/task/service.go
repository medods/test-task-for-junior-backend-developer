package task

import (
	"context"
	"fmt"
	"reflect"
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

	now := s.now()
	model := &taskdomain.Task{
		Title:          normalized.Title,
		Description:    normalized.Description,
		Status:         normalized.Status,
		Period:         normalized.Period,
		PeriodAnchorAt: periodAnchor(normalized.Period, now),
	}
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.enrichTask(created), nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.enrichTask(task), nil
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

	now := s.now()
	model := &taskdomain.Task{
		ID:             id,
		Title:          normalized.Title,
		Description:    normalized.Description,
		Status:         normalized.Status,
		Period:         normalized.Period,
		UpdatedAt:      now,
		PeriodAnchorAt: resolvePeriodAnchor(existing, normalized.Period, now),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.enrichTask(updated), nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		s.enrichTask(&tasks[i])
	}

	return tasks, nil
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

	normalizedPeriod, err := normalizePeriod(input.Period)
	if err != nil {
		return CreateInput{}, fmt.Errorf("%w: invalid period", ErrInvalidInput)
	}
	input.Period = normalizedPeriod

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

	normalizedPeriod, err := normalizePeriod(input.Period)
	if err != nil {
		return UpdateInput{}, fmt.Errorf("%w: invalid period", ErrInvalidInput)
	}
	input.Period = normalizedPeriod

	return input, nil
}

func normalizePeriod(period *taskdomain.Period) (*taskdomain.Period, error) {
	normalized, err := period.Normalize()
	if err != nil {
		return nil, err
	}

	return normalized, nil
}

func periodAnchor(period *taskdomain.Period, now time.Time) *time.Time {
	if period == nil {
		return nil
	}

	value := now.UTC()
	return &value
}

func resolvePeriodAnchor(existing *taskdomain.Task, period *taskdomain.Period, now time.Time) *time.Time {
	if period == nil {
		return nil
	}

	if existing != nil && reflect.DeepEqual(existing.Period, period) {
		if existing.PeriodAnchorAt != nil {
			value := existing.PeriodAnchorAt.UTC()
			return &value
		}

		value := existing.CreatedAt.UTC()
		return &value
	}

	return periodAnchor(period, now)
}

func (s *Service) enrichTask(task *taskdomain.Task) *taskdomain.Task {
	if task == nil {
		return nil
	}

	if task.Period == nil {
		task.ScheduledDate = nil
		return task
	}

	anchor := task.CreatedAt
	if task.PeriodAnchorAt != nil {
		anchor = *task.PeriodAnchorAt
	}
	task.ScheduledDate = task.Period.NextScheduledDate(s.now(), anchor)

	return task
}
