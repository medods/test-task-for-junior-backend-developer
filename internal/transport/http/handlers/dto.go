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
	Period      *periodDTO        `json:"period"`
}

type taskDTO struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	Period        *periodDTO        `json:"period,omitempty"`
	ScheduledDate *string           `json:"scheduled_date,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type periodDTO struct {
	Type          taskdomain.PeriodType   `json:"type"`
	IntervalDays  *int                    `json:"interval_days,omitempty"`
	MonthDay      *int                    `json:"month_day,omitempty"`
	SpecificDates []string                `json:"specific_dates,omitempty"`
	EvenOddType   *taskdomain.EvenOddType `json:"even_odd_type,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	var scheduledDate *string
	if task.ScheduledDate != nil {
		value := task.ScheduledDate.UTC().Format(time.DateOnly)
		scheduledDate = &value
	}

	return taskDTO{
		ID:            task.ID,
		Title:         task.Title,
		Description:   task.Description,
		Status:        task.Status,
		Period:        newPeriodDTO(task.Period),
		ScheduledDate: scheduledDate,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
	}
}

func (p *periodDTO) toDomain() (*taskdomain.Period, error) {
	if p == nil {
		return nil, nil
	}

	period := &taskdomain.Period{
		Type:         p.Type,
		IntervalDays: p.IntervalDays,
		MonthDay:     p.MonthDay,
		EvenOddType:  p.EvenOddType,
	}

	if len(p.SpecificDates) > 0 {
		period.SpecificDates = make([]time.Time, 0, len(p.SpecificDates))
		for _, raw := range p.SpecificDates {
			parsed, err := time.Parse(time.DateOnly, raw)
			if err != nil {
				return nil, fmt.Errorf("specific_dates must use YYYY-MM-DD")
			}

			period.SpecificDates = append(period.SpecificDates, parsed.UTC())
		}
	}

	return period, nil
}

func newPeriodDTO(period *taskdomain.Period) *periodDTO {
	if period == nil {
		return nil
	}

	dto := &periodDTO{
		Type:         period.Type,
		IntervalDays: period.IntervalDays,
		MonthDay:     period.MonthDay,
		EvenOddType:  period.EvenOddType,
	}
	if len(period.SpecificDates) > 0 {
		dto.SpecificDates = make([]string, 0, len(period.SpecificDates))
		for _, value := range period.SpecificDates {
			dto.SpecificDates = append(dto.SpecificDates, value.UTC().Format(time.DateOnly))
		}
	}

	return dto
}
