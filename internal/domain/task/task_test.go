package task

import (
	"testing"
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
)

func TestTaskActiveOn_NoneRecurrence(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC))
	task := Task{
		RecurrenceType:      recurrence.NoneRecurrence,
		RecurrenceStartDate: &start,
	}

	if !task.ActiveOn(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on start date")
	}
	if task.ActiveOn(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be inactive on a different date")
	}
}

func TestTaskActiveOn_DailyEveryN(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	task := Task{
		RecurrenceType: recurrence.DailyEveryN,
		RecurrencePayload: recurrence.RecurrencePayload{
			IntervalDays: 2,
		},
		RecurrenceStartDate: &start,
	}

	if !task.ActiveOn(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on start date")
	}
	if task.ActiveOn(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be inactive on day 1")
	}
	if !task.ActiveOn(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on day 2")
	}
}

func TestTaskActiveOn_MonthlyDay(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))
	task := Task{
		RecurrenceType: recurrence.MonthlyDay,
		RecurrencePayload: recurrence.RecurrencePayload{
			MonthDay: 15,
		},
		RecurrenceStartDate: &start,
	}

	if task.ActiveOn(time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be inactive on non-matching day")
	}
	if !task.ActiveOn(time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on matching day")
	}
	if !task.ActiveOn(time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on matching day next month")
	}
}

func TestTaskActiveOn_SpecificDates(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	d1 := common.NewDate(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC))
	d2 := common.NewDate(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))

	task := Task{
		RecurrenceType: recurrence.SpecificDates,
		RecurrencePayload: recurrence.RecurrencePayload{
			SpecificDates: []common.DateType{d1, d2},
		},
		RecurrenceStartDate: &start,
	}

	if !task.ActiveOn(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on a specific date")
	}
	if task.ActiveOn(time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be inactive on a non-specific date")
	}
	if !task.ActiveOn(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on a specific date")
	}
}

func TestTaskActiveOn_MonthParity(t *testing.T) {
	start := common.NewDate(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	task := Task{
		RecurrenceType: recurrence.MonthParity,
		RecurrencePayload: recurrence.RecurrencePayload{
			Parity: recurrence.Even,
		},
		RecurrenceStartDate: &start,
	}

	if task.ActiveOn(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be inactive on odd day")
	}
	if !task.ActiveOn(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected task to be active on even day")
	}
}
