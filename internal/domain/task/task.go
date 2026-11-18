package task

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type RecurrenceType string

const (
	RecurrenceDaily      RecurrenceType = "daily"       // каждый n-й день
	RecurrenceMonthlyDay RecurrenceType = "monthly_day" // каждое число месяца
	RecurrenceSpecific   RecurrenceType = "specific"    // конкретные даты
	RecurrenceEvenOdd    RecurrenceType = "even_odd"    // четные/нечетные дни месяца
)

type Status string
type TimeSlice []time.Time

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID               int64          `json:"id"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Status           Status         `json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	RecurrenceType   RecurrenceType `json:"recurrence_type"`
	RecurrenceEveryN *int           `json:"recurrence_every_n,omitempty"`                 // для daily
	MonthlyDay       *int           `json:"monthly_day,omitempty"`                        // 1-30 для monthly_day
	SpecificDates    TimeSlice      `json:"specific_dates,omitempty" db:"specific_dates"` // для specific
	EvenOdd          *string        `json:"even_odd,omitempty"`                           // "even" / "odd"
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// Scan реализует интерфейс sql.Scanner
func (ts *TimeSlice) Scan(value interface{}) error {
	if value == nil {
		*ts = TimeSlice{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan TimeSlice: unsupported type")
	}

	// Если строка пустая, возвращаем пустой слайс
	if len(bytes) == 0 || string(bytes) == "null" {
		*ts = TimeSlice{}
		return nil
	}

	var timeStrings []string
	if err := json.Unmarshal(bytes, &timeStrings); err != nil {
		return err
	}

	dates := make(TimeSlice, 0, len(timeStrings))
	for _, tsStr := range timeStrings {
		t, err := time.Parse(time.RFC3339, tsStr)
		if err != nil {
			// Пробуем другой формат
			t, err = time.Parse("2006-01-02", tsStr)
			if err != nil {
				return err
			}
		}
		dates = append(dates, t)
	}

	*ts = dates
	return nil
}

// Value реализует driver.Valuer
func (ts TimeSlice) Value() (driver.Value, error) {
	if len(ts) == 0 {
		return nil, nil
	}

	timeStrings := make([]string, len(ts))
	for i, t := range ts {
		timeStrings[i] = t.Format(time.RFC3339)
	}

	return json.Marshal(timeStrings)
}

func (t *Task) ValidateRecurrence() error {
	if t.RecurrenceType == "" {
		return nil // обычная задача без повторения
	}

	switch t.RecurrenceType {
	case RecurrenceDaily:
		if t.RecurrenceEveryN == nil || *t.RecurrenceEveryN < 1 {
			return errors.New("daily recurrence requires recurrence_every_n >= 1")
		}

	case RecurrenceMonthlyDay:
		if t.MonthlyDay == nil || *t.MonthlyDay < 1 || *t.MonthlyDay > 30 {
			return errors.New("monthly_day must be between 1 and 30")
		}

	case RecurrenceSpecific:
		if len(t.SpecificDates) == 0 {
			return errors.New("specific dates list cannot be empty")
		}

	case RecurrenceEvenOdd:
		if t.EvenOdd == nil || (*t.EvenOdd != "even" && *t.EvenOdd != "odd") {
			return errors.New("even_odd must be 'even' or 'odd'")
		}

	default:
		return errors.New("invalid recurrence_type: " + string(t.RecurrenceType))
	}

	return nil
}
