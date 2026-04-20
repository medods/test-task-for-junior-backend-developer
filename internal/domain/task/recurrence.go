package task

import (
	"errors"
	"fmt"
	"time"
)

type RecurrenceType string

const (
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specificDates"
	RecurrenceTypeParityDates   RecurrenceType = "parityDates"
)

type ParityDateType string

const (
	ParityDateTypeEven ParityDateType = "even"
	ParityDateTypeOdd  ParityDateType = "odd"
)

type RecurrenceConfig struct {
	Type          RecurrenceType `json:"type"` // daily, monthly, specific_dates, parity
	Daily         int            `json:"daily,omitempty"`
	Monthly       int            `json:"monthly,omitempty"`
	SpecificDates []string       `json:"specific_dates,omitempty"`
	ParityDates   ParityDateType `json:"parity_dates,omitempty"`
	RepeatCount   int            `json:"repeat_count,omitempty"`
}

func (r *RecurrenceConfig) Validate() error {
	if r == nil {
		return nil
	}

	switch r.Type {
	case RecurrenceTypeDaily:
		return r.validateDaily()
	case RecurrenceTypeMonthly:
		return r.validateMonthly()
	case RecurrenceTypeSpecificDates:
		return r.validateSpecificDates()
	case RecurrenceTypeParityDates:
		return r.validateParity()
	default:
		return fmt.Errorf("неизвестный тип повторения: %s", r.Type)
	}
}

func (r *RecurrenceConfig) validateDaily() error {
	if r.Daily < 1 || r.Daily > 31 {
		return errors.New("периодичность для DAILY не может быть меньше 1 или больше 31 дня")
	}
	return r.validateRepeatCount(365)
}

func (r *RecurrenceConfig) validateMonthly() error {
	if r.Monthly < 1 || r.Monthly > 31 {
		return errors.New("дата MONTHLY должна быть в диапазоне от 1 до 31")
	}
	return r.validateRepeatCount(24)
}

func (r *RecurrenceConfig) validateSpecificDates() error {
	count := len(r.SpecificDates)
	if count == 0 {
		return errors.New("SPECIFIC_DATES должна содержать хотя бы одну дату")
	}
	if count > 100 {
		return errors.New("количество дат в SPECIFIC_DATES не должно превышать 100")
	}
	for _, date := range r.SpecificDates {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return fmt.Errorf("неверный формат даты: %s, ожидался YYYY-MM-DD", date)
		}
	}
	return nil
}

func (r *RecurrenceConfig) validateParity() error {
	if r.ParityDates != ParityDateTypeEven && r.ParityDates != ParityDateTypeOdd {
		return errors.New("параметр parity_dates должен быть 'even' или 'odd'")
	}
	return r.validateRepeatCount(365)
}

func (r *RecurrenceConfig) validateRepeatCount(max int) error {
	if r.RepeatCount > max {
		return fmt.Errorf("количество создаваемых задач для типа %s не может превышать %d", r.Type, max)
	}
	return nil
}

func getDefaultRepeatCount(recurrenceType RecurrenceType) int { //установка деволтных значений
	switch recurrenceType {
	case RecurrenceTypeDaily, RecurrenceTypeParityDates:
		return 90
	case RecurrenceTypeMonthly:
		return 24
	case RecurrenceTypeSpecificDates:
		return 0 // будет использована длина массива
	default:
		return 0
	}
}
