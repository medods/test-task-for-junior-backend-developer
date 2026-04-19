package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/google/uuid"
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

	var dates []time.Time

	if input.Recurrence != nil {
		dates, err = generateDates(*input.Recurrence)
		if err != nil {
			return nil, err
		}
	}

	if len(dates) > 0 {
		tasks := make([]*taskdomain.Task, 0, len(dates))
		groupId := uuid.New().String()
		for _, date := range dates {
			now := s.now()
			tasks = append(tasks, &taskdomain.Task{
				Title:             normalized.Title,
				Description:       normalized.Description,
				Status:            normalized.Status,
				CreatedAt:         now,
				UpdatedAt:         now,
				RecurrenceType:    &normalized.Recurrence.Type,
				RecurrenceGroupID: &groupId,
				ScheduledDate:     &date,
			})
		}

		created, err := s.repo.CreateBatch(ctx, tasks)
		if err != nil {
			return nil, err
		}

		if len(created) > 0 {
			return created[0], nil
		}

		return nil, nil
	}

	now := s.now()

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
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

func (s *Service) DeleteByGroupID(ctx context.Context, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("%w: groupID must be not empty", ErrInvalidInput)
	}

	return s.repo.DeleteByGroupID(ctx, groupID)
}

func (s *Service) List(ctx context.Context, groupID string) ([]taskdomain.Task, error) {
	return s.repo.List(ctx, groupID)
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

	if input.Recurrence != nil {
		rec, err := validateRecurrenceInput(*input.Recurrence)
		if err != nil {
			return CreateInput{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		input.Recurrence = &rec
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

func validateRecurrenceInput(input RecurrenceInput) (RecurrenceInput, error) {
	if input.EndDate.Before(input.StartDate) {
		return RecurrenceInput{}, fmt.Errorf("%w: endDate is before startDate", ErrInvalidRecurrence)
	}

	if input.Type == daily && input.EveryNDays == nil {
		return RecurrenceInput{}, fmt.Errorf("%w: nDays not specified", ErrInvalidRecurrence)
	}

	if input.Type == monthly && input.DayOfMonth == nil {
		return RecurrenceInput{}, fmt.Errorf("%w: dayOfMonth not specified", ErrInvalidRecurrence)
	}

	if input.Type == monthly && (*input.DayOfMonth < 1 || *input.DayOfMonth > 31) {
		return RecurrenceInput{}, fmt.Errorf("%w: invalid dayOfMonth", ErrInvalidRecurrence)
	}

	if input.Type == evenOdd && input.Parity == nil {
		return RecurrenceInput{}, fmt.Errorf("%w: parity not specified", ErrInvalidRecurrence)
	}

	if input.Type == evenOdd && *input.Parity != "odd" && *input.Parity != "even" {
		return RecurrenceInput{}, fmt.Errorf("%w: invalid parity", ErrInvalidRecurrence)
	}

	if input.Type == specificDates && len(input.Dates) == 0 {
		return RecurrenceInput{}, fmt.Errorf("%w: dates not specified", ErrInvalidRecurrence)
	}

	return input, nil
}
