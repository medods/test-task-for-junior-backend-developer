package handlers

import (
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/shared/dateutil"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID               int64             `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Status           taskdomain.Status `json:"status"`
	ScheduledFor     string            `json:"scheduled_for"`
	RecurrenceRuleID *int64            `json:"recurrence_rule_id,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:               task.ID,
		Title:            task.Title,
		Description:      task.Description,
		Status:           task.Status,
		ScheduledFor:     dateutil.Format(task.ScheduledFor),
		RecurrenceRuleID: task.RecurrenceRuleID,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
}

type recurrenceRuleMutationDTO struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Active      *bool               `json:"active,omitempty"`
	Schedule    recurrenceConfigDTO `json:"schedule"`
}

type recurrenceConfigDTO struct {
	Type       recurrencedomain.ScheduleType `json:"type"`
	StartsOn   string                        `json:"starts_on,omitempty"`
	EndsOn     string                        `json:"ends_on,omitempty"`
	EveryNDays *int                          `json:"every_n_days,omitempty"`
	DayOfMonth *int                          `json:"day_of_month,omitempty"`
	DayParity  *recurrencedomain.DayParity   `json:"day_parity,omitempty"`
	Dates      []string                      `json:"dates,omitempty"`
}

type recurrenceRuleDTO struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Active      bool                `json:"active"`
	Schedule    recurrenceConfigDTO `json:"schedule"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func newRecurrenceRuleDTO(rule *recurrencedomain.Rule) recurrenceRuleDTO {
	dto := recurrenceRuleDTO{
		ID:          rule.ID,
		Title:       rule.Title,
		Description: rule.Description,
		Active:      rule.Active,
		Schedule: recurrenceConfigDTO{
			Type:       rule.ScheduleType,
			StartsOn:   dateutil.Format(rule.StartsOn),
			EndsOn:     formatOptionalDate(rule.EndsOn),
			EveryNDays: rule.EveryNDays,
			DayOfMonth: rule.DayOfMonth,
			DayParity:  rule.DayParity,
			Dates:      formatDates(rule.SpecificDates),
		},
		CreatedAt: rule.CreatedAt,
		UpdatedAt: rule.UpdatedAt,
	}

	return dto
}

func formatOptionalDate(value *time.Time) string {
	if value == nil {
		return ""
	}

	return dateutil.Format(*value)
}

func formatDates(values []time.Time) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, dateutil.Format(value))
	}

	return result
}
