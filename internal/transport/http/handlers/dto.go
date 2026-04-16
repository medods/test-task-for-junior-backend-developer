package handlers

import (
	"encoding/json"
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	task "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title               string                       `json:"title" validate:"required,len=128"`
	Description         string                       `json:"description" validate:"len=512"`
	Status              task.Status                  `json:"status" validate:"required,oneof=new in_progress done"`
	RecurrenceType      recurrence.RecurrenceType    `json:"recurrence_type" validate:"required"`
	RecurrencePayload   recurrence.RecurrencePayload `json:"recurrence_payload" validate:"required"`
	RecurrenceStartDate *common.DateType             `json:"recurrence_start_date" validate:"omitempty,datetime=2006-01-02"`
}

type taskPatchDTO struct {
	Title             *string                       `json:"title"`
	Description       *string                       `json:"description"`
	Status            *task.Status                  `json:"status"`
	RecurrenceType    *recurrence.RecurrenceType    `json:"recurrence_type"`
	RecurrencePayload *recurrence.RecurrencePayload `json:"recurrence_payload"`

	RecurrenceStartDate json.RawMessage `json:"recurrence_start_date"`
}

type taskDTO struct {
	ID                  int64                        `json:"id" validate:"required"`
	Title               string                       `json:"title" validate:"required,len=128"`
	Description         string                       `json:"description" validate:"len=512"`
	Status              task.Status                  `json:"status" validate:"required,oneof=new in_progress done"`
	CreatedAt           time.Time                    `json:"created_at" validate:"required"`
	UpdatedAt           time.Time                    `json:"updated_at" validate:"required"`
	RecurrenceType      recurrence.RecurrenceType    `json:"recurrence_type" validate:"required"`
	RecurrencePayload   recurrence.RecurrencePayload `json:"recurrence_payload" validate:"required"`
	RecurrenceStartDate *common.DateType             `json:"recurrence_start_date" validate:"omitempty,datetime=2006-01-02"`
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
