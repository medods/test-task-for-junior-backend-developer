// Package task
package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/taskservice/internal/domain/common"
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

func (s *Service) Patch(ctx context.Context, id int64, input PatchInput) (*task.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if input.Title == nil && input.Description == nil && input.Status == nil && input.RecurrenceType == nil && input.RecurrencePayload == nil && !input.RecurrenceStartDateSet {
		return nil, fmt.Errorf("%w: no fields to update", ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	upd := UpdateInput{
		Title:               existing.Title,
		Description:         existing.Description,
		Status:              existing.Status,
		RecurrenceType:      existing.RecurrenceType,
		RecurrencePayload:   existing.RecurrencePayload,
		RecurrenceStartDate: existing.RecurrenceStartDate,
	}

	if input.Title != nil {
		upd.Title = *input.Title
	}
	if input.Description != nil {
		upd.Description = *input.Description
	}
	if input.Status != nil {
		upd.Status = *input.Status
	}
	if input.RecurrenceType != nil {
		upd.RecurrenceType = *input.RecurrenceType
	}
	if input.RecurrencePayload != nil {
		upd.RecurrencePayload = *input.RecurrencePayload
	}
	if input.RecurrenceStartDateSet {
		upd.RecurrenceStartDate = input.RecurrenceStartDate
	}

	normalized, err := validateUpdateInput(upd)
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
	normalized := common.NewDate(date.UTC()).Time
	return s.repo.ListByDate(ctx, normalized)
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

	if err := validateRecurrence(input.RecurrenceType, input.RecurrencePayload, input.RecurrenceStartDate); err != nil {
		return CreateInput{}, err
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

	if err := validateRecurrence(input.RecurrenceType, input.RecurrencePayload, input.RecurrenceStartDate); err != nil {
		return UpdateInput{}, err
	}

	return input, nil
}

func validateRecurrence(rt recurrence.RecurrenceType, payload recurrence.RecurrencePayload, startDate *common.DateType) error {
	switch rt {
	case recurrence.NoneRecurrence,
		recurrence.DailyEveryN,
		recurrence.MonthlyDay,
		recurrence.SpecificDates,
		recurrence.MonthParity:
	default:
		return fmt.Errorf("%w: invalid recurrence_type", ErrInvalidInput)
	}

	if rt != recurrence.NoneRecurrence && startDate == nil {
		return fmt.Errorf("%w: recurrence_start_date is required", ErrInvalidInput)
	}

	switch rt {
	case recurrence.NoneRecurrence:
		return nil
	case recurrence.DailyEveryN:
		if payload.IntervalDays < 1 {
			return fmt.Errorf("%w: interval_days must be >= 1", ErrInvalidInput)
		}
		return nil
	case recurrence.MonthlyDay:
		if payload.MonthDay < 1 || payload.MonthDay > 30 {
			return fmt.Errorf("%w: month_day must be between 1 and 30", ErrInvalidInput)
		}
		return nil
	case recurrence.SpecificDates:
		if len(payload.SpecificDates) == 0 {
			return fmt.Errorf("%w: specific_dates must not be empty", ErrInvalidInput)
		}

		start := common.NewDate(startDate.Time.UTC()).Time
		seen := make(map[string]struct{}, len(payload.SpecificDates))
		for _, sd := range payload.SpecificDates {
			d := common.NewDate(sd.Time.UTC()).Time
			if d.Before(start) {
				return fmt.Errorf("%w: specific_dates must be on/after recurrence_start_date", ErrInvalidInput)
			}

			key := d.Format(common.DateLayout)
			if _, ok := seen[key]; ok {
				return fmt.Errorf("%w: specific_dates must be unique", ErrInvalidInput)
			}
			seen[key] = struct{}{}
		}

		return nil
	case recurrence.MonthParity:
		if payload.Parity != recurrence.Even && payload.Parity != recurrence.Odd {
			return fmt.Errorf("%w: parity must be even or odd", ErrInvalidInput)
		}
		return nil
	default:
		return fmt.Errorf("%w: invalid recurrence_type", ErrInvalidInput)
	}
}
