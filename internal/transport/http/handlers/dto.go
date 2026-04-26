package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	RecurrenceID *int64            `json:"recurrence_id,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		RecurrenceID: task.RecurrenceID,
	}
}

type recurrenceRequestDTO struct {
	Type       string   `json:"type"`
	Interval   *int     `json:"interval,omitempty"`
	DayOfMonth *int     `json:"day_of_month,omitempty"`
	WeekDays   []int    `json:"week_days,omitempty"`
	Dates      []string `json:"dates,omitempty"`
}

type recurrenceResponseDTO struct {
	ID               int64    `json:"id"`
	Type             string   `json:"type"`
	Interval         *int     `json:"interval,omitempty"`
	DayOfMonth       *int     `json:"day_of_month,omitempty"`
	WeekDays         []int    `json:"week_days,omitempty"`
	Dates            []string `json:"dates,omitempty"`
	NextOccurrenceAt string   `json:"next_occurrence_at"`
}

func recurrenceToDTO(id int64, rec taskdomain.Recurrence, nextAt time.Time) recurrenceResponseDTO {
	dto := recurrenceResponseDTO{
		ID:               id,
		Type:             rec.RecurrenceType(),
		NextOccurrenceAt: nextAt.Format("2006-01-02"),
	}

	switch r := rec.(type) {
	case taskdomain.DailyRecurrence:
		dto.Interval = &r.Interval
	case taskdomain.MonthlyRecurrence:
		dto.DayOfMonth = &r.DayOfMonth
	case taskdomain.WeeklyRecurrence:
		dto.WeekDays = r.WeekDays
	case taskdomain.SpecificDatesRecurrence:
		dto.Dates = r.Dates
	}

	return dto
}

func recurrenceFromDTO(req recurrenceRequestDTO) (taskdomain.Recurrence, error) {
	switch req.Type {
	case "daily":
		if req.Interval == nil {
			return nil, errMissingField("interval")
		}
		return taskdomain.DailyRecurrence{Interval: *req.Interval}, nil
	case "monthly":
		if req.DayOfMonth == nil {
			return nil, errMissingField("day_of_month")
		}
		return taskdomain.MonthlyRecurrence{DayOfMonth: *req.DayOfMonth}, nil
	case "weekly":
		return taskdomain.WeeklyRecurrence{WeekDays: req.WeekDays}, nil
	case "specific_dates":
		return taskdomain.SpecificDatesRecurrence{Dates: req.Dates}, nil
	default:
		return nil, errUnknownType(req.Type)
	}
}

func errMissingField(field string) error {
	return &validationError{msg: "missing required field: " + field}
}

func errUnknownType(t string) error {
	return &validationError{msg: "unknown recurrence type: " + t}
}

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
