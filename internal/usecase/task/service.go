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

	now := s.now()
	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		ScheduledAt: normalized.ScheduledAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.repo.Create(ctx, model)
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
		ScheduledAt: normalized.ScheduledAt,
		UpdatedAt:   s.now(),
	}

	return s.repo.Update(ctx, model)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]taskdomain.Task, error) {
	return s.repo.List(ctx, filter)
}

// CreateRecurring expands the recurrence rule into a set of dates, then creates
// one task per date. All tasks share a recurrence_group_id so they can be
// queried and managed together.
func (s *Service) CreateRecurring(ctx context.Context, input CreateRecurringInput) ([]taskdomain.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}
	if !input.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	dates, err := taskdomain.ExpandDates(input.Recurrence)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}
	if len(dates) == 0 {
		return nil, fmt.Errorf("%w: recurrence produces no dates in the given range", ErrInvalidInput)
	}

	groupID := taskdomain.NewGroupID()
	recurrenceCfg := input.Recurrence.Config()
	now := s.now()

	tasks := make([]*taskdomain.Task, 0, len(dates))
	for _, d := range dates {
		d := d // capture
		tasks = append(tasks, &taskdomain.Task{
			Title:             title,
			Description:       strings.TrimSpace(input.Description),
			Status:            input.Status,
			ScheduledAt:       &d,
			RecurrenceGroupID: &groupID,
			Recurrence:        recurrenceCfg,
			CreatedAt:         now,
			UpdatedAt:         now,
		})
	}

	return s.repo.CreateBatch(ctx, tasks)
}

// DeleteByGroupID removes every task belonging to groupID.
func (s *Service) DeleteByGroupID(ctx context.Context, groupID string) (int64, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return 0, fmt.Errorf("%w: group_id is required", ErrInvalidInput)
	}

	return s.repo.DeleteByGroupID(ctx, groupID)
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
