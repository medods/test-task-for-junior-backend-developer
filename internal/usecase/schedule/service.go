package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

type Service struct {
	repo        Repository
	txManager   TxManager
	materialize WindowMaterializer
	location    *time.Location
	windowDays  int
	now         func() time.Time
}

func NewService(repo Repository, txManager TxManager, materialize WindowMaterializer, location *time.Location, windowDays int) *Service {
	return &Service{
		repo:        repo,
		txManager:   txManager,
		materialize: materialize,
		location:    location,
		windowDays:  windowDays,
		now:         time.Now,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*scheduledomain.TaskSchedule, error) {
	model, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	var created *scheduledomain.TaskSchedule
	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		created, err = s.repo.Create(txCtx, model)
		if err != nil {
			return err
		}

		if !created.Active {
			return nil
		}

		from, to := s.defaultWindow()
		return s.materialize.MaterializeScheduleWindow(txCtx, created.ID, from, to)
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*scheduledomain.TaskSchedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]scheduledomain.TaskSchedule, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.TaskSchedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	model, err := validateUpdateInput(id, input)
	if err != nil {
		return nil, err
	}

	var updated *scheduledomain.TaskSchedule
	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		updated, err = s.repo.Update(txCtx, model)
		if err != nil {
			return err
		}

		from, to := s.defaultWindow()
		return s.materialize.RebuildScheduleWindow(txCtx, updated.ID, from, to)
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Deactivate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Deactivate(ctx, id, s.now().UTC())
}

func (s *Service) defaultWindow() (time.Time, time.Time) {
	from := scheduledomain.DateFromLocation(s.now(), s.location)
	return from, from.AddDate(0, 0, s.windowDays)
}

func validateCreateInput(input CreateInput) (*scheduledomain.TaskSchedule, error) {
	active := true
	if input.Active != nil {
		active = *input.Active
	}

	model := &scheduledomain.TaskSchedule{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Rule:        input.Rule,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Active:      active,
	}
	model.Normalize()

	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	return model, nil
}

func validateUpdateInput(id int64, input UpdateInput) (*scheduledomain.TaskSchedule, error) {
	active := true
	if input.Active != nil {
		active = *input.Active
	}

	model := &scheduledomain.TaskSchedule{
		ID:          id,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Rule:        input.Rule,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Active:      active,
	}
	model.Normalize()

	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}

	return model, nil
}
