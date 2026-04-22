package task

import (
	"context"
	"errors"
	"example.com/taskservice/internal/domain"
	"example.com/taskservice/internal/repository"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type CreateInput struct {
	Title       string
	Description string
	Status      domain.Status
	TaskDate    time.Time
	Recurrence  *RecurrenceInput
}

type RecurrenceInput struct {
	Type   domain.RecurrenceType
	Config domain.RecurrenceConfig
}
type UpdateInput struct {
	Title       string
	Description string
	Status      domain.Status
}

type ServiceAPI interface {
	Create(ctx context.Context, input CreateInput) (*domain.Task, error)
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*domain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*domain.Task, error)
	DeactivateTemplate(ctx context.Context, id int64) error
	ActivateTemplate(ctx context.Context, id int64) error
}

var ErrInvalidInput = errors.New("invalid task input")

type Service struct {
	repo   repository.Repository
	now    func() time.Time
	logger *slog.Logger
}

func NewService(repo repository.Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		now:    func() time.Time { return time.Now().UTC() },
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	if normalized.Recurrence != nil {
		return s.handleTemplateCreation(ctx, normalized)
	}

	model := &domain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		TaskDate:    normalized.TaskDate,
		CreatedAt:   s.now(),
		UpdatedAt:   s.now(),
	}

	return s.repo.Create(ctx, model)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*domain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &domain.Task{
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

func (s *Service) List(ctx context.Context) ([]*domain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) DeactivateTemplate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.UpdateTemplateStatus(ctx, id, false)
}

func (s *Service) ActivateTemplate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.UpdateTemplateStatus(ctx, id, true)
}

func (s *Service) ProcessTemplates(ctx context.Context) {

	templates, err := s.repo.GetActiveTemplates(ctx)
	if err != nil {
		s.logger.Error("failed to get active templates: error", err)
		return
	}

	now := s.now()
	startDate := now
	endDate := now.AddDate(0, 1, 0)

	for _, t := range templates {
		dates := calculateDates(t.RecurrenceConfig, t.RecurrenceType, startDate, endDate)

		for _, d := range dates {
			task := &domain.Task{
				TemplateID:  t.ID,
				Title:       t.Title,
				Description: t.Description,
				Status:      domain.StatusNew,
				TaskDate:    d,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			_, err = s.repo.Create(ctx, task)
			if err != nil {
				continue
			}
		}
	}
}

func (s *Service) saveTasks(ctx context.Context, title, desc string, templateID int64, dates []time.Time) (*domain.Task, error) {
	var firstTask *domain.Task

	for _, date := range dates {
		task := &domain.Task{
			TemplateID:  templateID,
			Title:       title,
			Description: desc,
			TaskDate:    date,
			Status:      domain.StatusNew,
			CreatedAt:   s.now(),
			UpdatedAt:   s.now(),
		}

		t, err := s.repo.Create(ctx, task)
		if err != nil {
			s.logger.Error("failed to create task", "error", err, "date", date, "template_id", templateID)
			continue
		}

		if firstTask == nil {
			firstTask = t
		}
	}

	if firstTask == nil {
		return nil, fmt.Errorf("no tasks were created")
	}

	return firstTask, nil
}

func (s *Service) handleTemplateCreation(ctx context.Context, input CreateInput) (*domain.Task, error) {
	input.Recurrence.Config.Hour = input.TaskDate.Hour()
	input.Recurrence.Config.Minute = input.TaskDate.Minute()

	startDate := s.now()
	endDate := startDate.AddDate(0, 1, 0)
	dates := calculateDates(input.Recurrence.Config, input.Recurrence.Type, startDate, endDate)

	if len(dates) == 0 {
		return nil, fmt.Errorf("invalid recurrence: no tasks generated for the period")
	}

	if input.Recurrence.Type == domain.Dates {
		return s.saveTasks(ctx, input.Title, input.Description, 0, dates)
	}

	templateModel := &domain.Template{
		Title:            input.Title,
		Description:      input.Description,
		RecurrenceType:   input.Recurrence.Type,
		RecurrenceConfig: input.Recurrence.Config,
		IsActive:         true,
		CreatedAt:        s.now(),
		UpdatedAt:        s.now(),
	}

	createdTemplate, err := s.repo.CreateTemplate(ctx, templateModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return s.saveTasks(ctx, input.Title, input.Description, createdTemplate.ID, dates)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = domain.StatusNew
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

func calculateDates(config domain.RecurrenceConfig, rType domain.RecurrenceType, startDate, endDate time.Time) []time.Time {
	var dates []time.Time

	baseStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)

	for current := baseStart; !current.After(end); current = current.AddDate(0, 0, 1) {

		createWithTime := func(d time.Time) time.Time {
			return time.Date(d.Year(), d.Month(), d.Day(), config.Hour, config.Minute, 0, 0, time.UTC)
		}

		switch rType {
		case domain.Daily:
			if config.Interval <= 0 {
				continue
			}
			daysDiff := int(current.Sub(baseStart).Hours() / 24)
			if daysDiff%config.Interval == 0 {
				dates = append(dates, createWithTime(current))
			}

		case domain.Monthly:
			if config.Day <= 0 {
				continue
			}
			lastDay := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			targetDay := config.Day
			if targetDay > lastDay {
				targetDay = lastDay
			}
			if current.Day() == targetDay {
				dates = append(dates, createWithTime(current))
			}

		case domain.Parity:
			isEven := current.Day()%2 == 0
			if (config.Parity == "even" && isEven) || (config.Parity == "odd" && !isEven) {
				dates = append(dates, createWithTime(current))
			}

		case domain.Dates:
			for _, d := range config.Dates {
				if d.Year() == current.Year() && d.Month() == current.Month() && d.Day() == current.Day() {
					dates = append(dates, createWithTime(current))
				}
			}
		}
	}
	return dates
}
