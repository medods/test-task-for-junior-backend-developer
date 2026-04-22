package task

import (
	"time"
) // Импортируем rrule для работы с периодичностью

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID           int64        `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       Status       `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Periodicity  *Periodicity `json:"periodicity,omitempty"`   // Настройки периодичности (если есть)
	ScheduledFor *time.Time   `json:"scheduled_for,omitempty"` // Дата экземпляра задачи (если это автосозданная задача)
	ParentID     *int64       `json:"parent_id,omitempty"`     // ID родительской задачи, если это экземпляр
	LastRunAt    *time.Time   `json:"last_run_at,omitempty"`   // Последний запуск воркера для периодической задачи
}

// Periodicity описывает правила повторения задачи по стандарту RFC 5545 (RRule)
type Periodicity struct {
	RRule   string   `json:"rrule"`             // RRULE-строка, например: "FREQ=WEEKLY;BYDAY=MO,WE,FR"
	ExDates []string `json:"exdates,omitempty"` // Исключённые даты (YYYY-MM-DD)
	RDates  []string `json:"rdates,omitempty"`  // Дополнительные даты
}

// Пример: создание периодичности для задачи
// rruleStr := "FREQ=DAILY;INTERVAL=2" // Каждый второй день
// periodicity := &Periodicity{RRule: rruleStr}

// Для генерации дат используйте стратегии из пакета periodicity

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
