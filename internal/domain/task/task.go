package task

import (
	"errors"
	"fmt"
	"time"
)

// Настройка периодичности повторений
type RecurrenceConfig struct {
	Type          string   `json:"type"` // daily, monthly, specific_dates, parity
	Daily         int      `json:"daily,omitempty"`
	Monthly       int      `json:"monthly,omitempty"`
	SpecificDates []string `json:"specific_dates,omitempty"`
	ParityDates   string   `json:"parity_dates,omitempty"`
	RepeatCount   int      `json:"repeat_count,omitempty"`
}

func (r *RecurrenceConfig) Validate() error {
	if r == nil {
		return nil
	}

	switch r.Type {
	case "daily":
		if r.Daily < 1 {
			return errors.New("Переодичность для ежедневных задач не может быть меньше 1 дня")
		}
		if r.Daily > 30 {
			return errors.New("Переодичность для ежедневных задач не может превышать 30 дней")
		}
		if r.RepeatCount > 365 {
			return errors.New("Количество создаваемых задач не может превышать 365")
		}

	case "monthly":
		if r.Monthly < 1 || r.Monthly > 30 {
			return errors.New("Дата ежемесячной задачи должна быть в диапозоне от 1 до 30")
		}
		if r.RepeatCount > 24 {
			return errors.New("Количество создаваемых задач не может превышать 24")
		}

	case "specific_dates":
		if len(r.SpecificDates) == 0 {
			return errors.New("Задача на конкретные даты не должена не иметь дат")
		}
		if len(r.SpecificDates) > 100 {
			return errors.New("Количсетво дат в задаче не должно превышать 100")
		}
		for _, date := range r.SpecificDates {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return fmt.Errorf("Неверный формат даты: %s, ожидался формат YYYY-MM-DD", date)
			}
		}

	case "parity":
		if r.ParityDates != "even" && r.ParityDates != "odd" {
			return errors.New("Параметр parity_type должен быть 'even' или 'odd'")
		}
		if r.RepeatCount > 365 {
			return errors.New("Количество создаваемых задач не может превышать 365")
		}

	default:
		return fmt.Errorf("Неизвестный тип повторения: %s", r.Type)
	}

	return nil
}

func getDefaultRepeatCount(recurrenceType string) int { //установка деволтных значений
	switch recurrenceType {
	case "daily", "parity":
		return 90
	case "monthly":
		return 24
	case "specific_dates":
		return 0 // будет использована длина массива
	default:
		return 0
	}
}

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID                  int64             `json:"id"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Status              Status            `json:"status"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	ScheduledAt         time.Time         `json:"scheduled_at"`
	Recurrence          *RecurrenceConfig `json:"recurrence,omitempty"`
	IsRecurringTemplate bool              `json:"is_recurring_template"`
	ParentTemplateID    *int64            `json:"parent_template_id,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
