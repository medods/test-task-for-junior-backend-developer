package periodicity

import "time"

// Интерфейс стратегии генерации дат для периодичности задачи
// Позволяет реализовать разные типы повторений (rrule, cron, custom и т.д.)
type Strategy interface {
	Occurrences(from, to time.Time) ([]time.Time, error)
	NextAfter(after time.Time) (time.Time, error)
	Type() string
}
