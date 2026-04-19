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

	now := s.now()

	if normalized.Recurrence == nil {
		return s.repo.Create(ctx, &taskdomain.Task{
			Title:       normalized.Title,
			Description: normalized.Description,
			Status:      normalized.Status,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	dates, err := generateDates(*normalized.Recurrence)
	if err != nil {
		return nil, err
	}
	if len(dates) == 0 {
		return nil, nil
	}

	groupID := uuid.New().String()
	tasks := make([]*taskdomain.Task, len(dates))
	for i, date := range dates {
		tasks[i] = &taskdomain.Task{
			Title:             normalized.Title,
			Description:       normalized.Description,
			Status:            normalized.Status,
			CreatedAt:         now,
			UpdatedAt:         now,
			RecurrenceType:    &normalized.Recurrence.Type,
			RecurrenceGroupID: &groupID,
			ScheduledDate:     &date,
		}
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
		ID:            id,
		Title:         normalized.Title,
		Description:   normalized.Description,
		Status:        normalized.Status,
		ScheduledDate: normalized.ScheduledDate,
		UpdatedAt:     s.now(),
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

	inputType := RecurrenceType(input.Type)

	if inputType == Daily && input.EveryNDays == nil {
		return RecurrenceInput{}, fmt.Errorf("%w: nDays not specified", ErrInvalidRecurrence)
	}

	if inputType == Monthly && input.DayOfMonth == nil {
		return RecurrenceInput{}, fmt.Errorf("%w: dayOfMonth not specified", ErrInvalidRecurrence)
	}

	if inputType == Monthly && (*input.DayOfMonth < 1 || *input.DayOfMonth > 31) {
		return RecurrenceInput{}, fmt.Errorf("%w: invalid dayOfMonth", ErrInvalidRecurrence)
	}

	if inputType == EvenOdd && input.Parity == nil {
		return RecurrenceInput{}, fmt.Errorf("%w: parity not specified", ErrInvalidRecurrence)
	}

	if inputType == EvenOdd && *input.Parity != string(Odd) && *input.Parity != string(Even) {
		return RecurrenceInput{}, fmt.Errorf("%w: invalid parity", ErrInvalidRecurrence)
	}

	if inputType == SpecificDates && len(input.Dates) == 0 {
		return RecurrenceInput{}, fmt.Errorf("%w: dates not specified", ErrInvalidRecurrence)
	}

	return input, nil
}
