package handlers

import (
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	task "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title               string                       `json:"title" validate:"required,max=128"`
	Description         string                       `json:"description" validate:"max=512"`
	Status              task.Status                  `json:"status" validate:"omitempty,oneof=new in_progress done"`
	RecurrenceType      recurrence.RecurrenceType    `json:"recurrence_type" validate:"required,oneof=none daily_every_n monthly_day specific_dates month_parity"`
	RecurrencePayload   recurrence.RecurrencePayload `json:"recurrence_payload"`
	RecurrenceStartDate *common.DateType             `json:"recurrence_start_date"`
}

type taskDTO struct {
	ID                  int64                        `json:"id" validate:"required"`
	Title               string                       `json:"title" validate:"required,max=128"`
	Description         string                       `json:"description" validate:"max=512"`
	Status              task.Status                  `json:"status" validate:"required,oneof=new in_progress done"`
	CreatedAt           time.Time                    `json:"created_at" validate:"required"`
	UpdatedAt           time.Time                    `json:"updated_at" validate:"required"`
	RecurrenceType      recurrence.RecurrenceType    `json:"recurrence_type" validate:"required,oneof=none daily_every_n monthly_day specific_dates month_parity"`
	RecurrencePayload   recurrence.RecurrencePayload `json:"recurrence_payload"`
	RecurrenceStartDate *common.DateType             `json:"recurrence_start_date"`
}

func newTaskDTO(task *task.Task) taskDTO {
	return taskDTO{
		ID:                  task.ID,
		Title:               task.Title,
		Description:         task.Description,
		Status:              task.Status,
		CreatedAt:           task.CreatedAt,
		UpdatedAt:           task.UpdatedAt,
		RecurrenceType:      task.RecurrenceType,
		RecurrencePayload:   task.RecurrencePayload,
		RecurrenceStartDate: task.RecurrenceStartDate,
	}
}
