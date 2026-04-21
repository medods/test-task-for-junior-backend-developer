package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Recurrence  *Recurrence `json:"recurrence,omitempty"` // может быть nil, если задача одноразовая
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

type RecurrenceType string

const (
	RecurrenceEveryNDays    RecurrenceType = "every_n_days"
	RecurrenceMonthlyDay    RecurrenceType = "monthly_day"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenOddDays   RecurrenceType = "even_odd_days"
)

type Recurrence struct {
	Type          RecurrenceType `json:"type"`                     // тип периодичности
	IntervalDays  *int           `json:"interval_days,omitempty"`  // для RecurrenceEveryNDays
	MonthDay      *int           `json:"month_day,omitempty"`      // для RecurrenceMonthlyDay
	SpecificDates []time.Time    `json:"specific_dates,omitempty"` // для RecurrenceSpecificDates
	EvenDays      *bool          `json:"even_days,omitempty"`      // для RecurrenceEvenOddDays
}

// Valid проверяет корректность типа RecurrenceType при создании/обновлении задачи
func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceEveryNDays, RecurrenceMonthlyDay, RecurrenceSpecificDates, RecurrenceEvenOddDays:
		return true
	default:
		return false
	}
}

// Valid проверяет заполнение поля структуры Recurrence для соответствующего типа рекуррентности Type
func (r *Recurrence) Valid() bool {
	if !r.Type.Valid() {
		return false
	}

	switch r.Type {
	case RecurrenceEveryNDays:
		return r.IntervalDays != nil && 0 < *r.IntervalDays
	case RecurrenceMonthlyDay:
		return r.MonthDay != nil && 1 <= *r.MonthDay && *r.MonthDay <= 31 // допускаем 31 день в месяце
	case RecurrenceSpecificDates:
		return len(r.SpecificDates) > 0
	case RecurrenceEvenOddDays:
		return r.EvenDays != nil
	}

	/*
		Тут может быть проверка на строгое правило:
		"все поля, не относящиеся к r.Type, должны быть nil"
		(поле SpecificDates в текущей реализации - просто пустой слайс)
	*/

	return false
}
