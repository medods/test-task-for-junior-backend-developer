package task

import (
	"context"
	"fmt"
	"slices"
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
		Recurrence:  normalized.Recurrence,
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
		Recurrence:  normalized.Recurrence,
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

	recurrence, err := validateRecurrence(input.Recurrence)
	if err != nil {
		return CreateInput{}, err
	}
	input.Recurrence = recurrence

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

	recurrence, err := validateRecurrence(input.Recurrence)
	if err != nil {
		return UpdateInput{}, err
	}
	input.Recurrence = recurrence

	return input, nil
}

func validateRecurrence(recurrence taskdomain.Recurrence) (taskdomain.Recurrence, error) {
	// Пустой recurrence в запросе трактуется как отсутствие правил периодичности.
	if recurrence.Type == "" {
		recurrence.Type = taskdomain.RecurrenceTypeNone
	}

	if !recurrence.Type.Valid() {
		return taskdomain.Recurrence{}, fmt.Errorf("%w: invalid recurrence.type", ErrInvalidInput)
	}

	switch recurrence.Type {
	case taskdomain.RecurrenceTypeNone:
		if recurrence.EveryNDays != nil || recurrence.DayOfMonth != nil || recurrence.Parity != nil || len(recurrence.Dates) > 0 {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: recurrence fields must be empty for type none", ErrInvalidInput)
		}
	case taskdomain.RecurrenceTypeDaily:
		if recurrence.EveryNDays == nil || *recurrence.EveryNDays < 1 || *recurrence.EveryNDays > 365 {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: recurrence.every_n_days must be in range 1..365", ErrInvalidInput)
		}
		recurrence.DayOfMonth = nil
		recurrence.Parity = nil
		recurrence.Dates = nil
	case taskdomain.RecurrenceTypeMonthly:
		if recurrence.DayOfMonth == nil || *recurrence.DayOfMonth < 1 || *recurrence.DayOfMonth > 30 {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: recurrence.day_of_month must be in range 1..30", ErrInvalidInput)
		}
		recurrence.EveryNDays = nil
		recurrence.Parity = nil
		recurrence.Dates = nil
	case taskdomain.RecurrenceTypeSpecificDates:
		if len(recurrence.Dates) == 0 {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: recurrence.dates must contain at least one date", ErrInvalidInput)
		}

		// Нормализуем список дат для детерминированного хранения:
		// удаляем лишние пробелы, валидируем формат, убираем дубликаты и сортируем.
		uniqueDates := make(map[string]struct{}, len(recurrence.Dates))
		validatedDates := make([]string, 0, len(recurrence.Dates))
		for _, rawDate := range recurrence.Dates {
			date := strings.TrimSpace(rawDate)
			if date == "" {
				return taskdomain.Recurrence{}, fmt.Errorf("%w: recurrence.dates contains empty value", ErrInvalidInput)
			}
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return taskdomain.Recurrence{}, fmt.Errorf("%w: recurrence.dates must use YYYY-MM-DD format", ErrInvalidInput)
			}
			if _, exists := uniqueDates[date]; !exists {
				uniqueDates[date] = struct{}{}
				validatedDates = append(validatedDates, date)
			}
		}
		slices.Sort(validatedDates)
		recurrence.Dates = validatedDates
		recurrence.EveryNDays = nil
		recurrence.DayOfMonth = nil
		recurrence.Parity = nil
	case taskdomain.RecurrenceTypeMonthParity:
		if recurrence.Parity == nil || !recurrence.Parity.Valid() {
			return taskdomain.Recurrence{}, fmt.Errorf("%w: recurrence.parity must be even or odd", ErrInvalidInput)
		}
		recurrence.EveryNDays = nil
		recurrence.DayOfMonth = nil
		recurrence.Dates = nil
	default:
		return taskdomain.Recurrence{}, fmt.Errorf("%w: unsupported recurrence.type", ErrInvalidInput)
	}

	return recurrence, nil
}
