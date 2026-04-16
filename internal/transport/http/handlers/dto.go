package handlers

import (
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	ScheduledAt *time.Time        `json:"scheduled_at"`
}

type taskDTO struct {
	ID                int64                        `json:"id"`
	Title             string                       `json:"title"`
	Description       string                       `json:"description"`
	Status            taskdomain.Status            `json:"status"`
	ScheduledAt       *time.Time                   `json:"scheduled_at,omitempty"`
	RecurrenceGroupID *string                      `json:"recurrence_group_id,omitempty"`
	Recurrence        *taskdomain.RecurrenceConfig `json:"recurrence,omitempty"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:                task.ID,
		Title:             task.Title,
		Description:       task.Description,
		Status:            task.Status,
		ScheduledAt:       task.ScheduledAt,
		RecurrenceGroupID: task.RecurrenceGroupID,
		Recurrence:        task.Recurrence,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
	}
}

type recurrenceDTO struct {
	Type        string   `json:"type"`
	EveryNDays  int      `json:"every_n_days"`
	DaysOfMonth []int    `json:"days_of_month"`
	Parity      string   `json:"parity"`
	Dates       []string `json:"dates"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
}

type createRecurringRequestDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  recurrenceDTO     `json:"recurrence"`
}

type createRecurringResponseDTO struct {
	RecurrenceGroupID string    `json:"recurrence_group_id"`
	Tasks             []taskDTO `json:"tasks"`
}

func (dto recurrenceDTO) toRecurrenceInput() (taskdomain.RecurrenceInput, error) {
	const dateFmt = "2006-01-02"

	input := taskdomain.RecurrenceInput{
		Type:        taskdomain.RecurrenceType(dto.Type),
		EveryNDays:  dto.EveryNDays,
		DaysOfMonth: dto.DaysOfMonth,
		Parity:      dto.Parity,
	}

	if dto.StartDate != "" {
		t, err := time.Parse(dateFmt, dto.StartDate)
		if err != nil {
			return taskdomain.RecurrenceInput{}, fmt.Errorf("invalid start_date %q: use YYYY-MM-DD format", dto.StartDate)
		}
		input.StartDate = t
	}

	if dto.EndDate != "" {
		t, err := time.Parse(dateFmt, dto.EndDate)
		if err != nil {
			return taskdomain.RecurrenceInput{}, fmt.Errorf("invalid end_date %q: use YYYY-MM-DD format", dto.EndDate)
		}
		input.EndDate = t
	}

	for _, ds := range dto.Dates {
		t, err := time.Parse(dateFmt, ds)
		if err != nil {
			return taskdomain.RecurrenceInput{}, fmt.Errorf("invalid date %q in dates: use YYYY-MM-DD format", ds)
		}
		input.Dates = append(input.Dates, t)
	}

	return input, nil
}
