package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)
type RecurrenceDTO struct {
	Type 	 		string 	 `json:"type"`	
	Interval 		*int    	 `json:"interval,omitempty"`
	DayOfMonth 		*int    	 `json:"day_of_month,omitempty"`
	SpecificDates 	[]string `json:"specific_dates,omitempty"`
	IsEven 			*bool 	 `json:"is_even,omitempty"`
}

type taskMutationDTO struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	RecurrenceDTO *RecurrenceDTO    `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          	int64             `json:"id"`
	Title       	string            `json:"title"`
	Description 	string            `json:"description"`
	Status      	taskdomain.Status `json:"status"`
	CreatedAt   	time.Time         `json:"created_at"`
	UpdatedAt   	time.Time         `json:"updated_at"`
	RecurrenceDTO   *RecurrenceDTO    `json:"recurrence,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,

	}

	if task.Recurrence != nil { 
		recurrenceDTO := &RecurrenceDTO{
			Type: string(task.Recurrence.Type),
			Interval: task.Recurrence.Interval,
			DayOfMonth: task.Recurrence.DayOfMonth,
			IsEven: task.Recurrence.IsEven,
		}
		if len(task.Recurrence.SpecificDates) > 0 {
			recurrenceDTO.SpecificDates = make([]string, len(task.Recurrence.SpecificDates))
			for i, d := range task.Recurrence.SpecificDates{ 
				recurrenceDTO.SpecificDates[i] = d.Format("2006-01-02")
			} 
		}
		dto.RecurrenceDTO = recurrenceDTO
	}
	return dto

}
