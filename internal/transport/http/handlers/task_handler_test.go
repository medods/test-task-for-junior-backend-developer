package handlers

import (
	"net/http/httptest"
	"testing"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

func TestDecodeTaskListInputParsesAllFilters(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("GET", "/api/v1/tasks?date_from=2026-04-15&date_to=2026-05-15&schedule_id=42", nil)

	filter, err := decodeTaskListInput(request)
	if err != nil {
		t.Fatalf("decode task list input: %v", err)
	}

	if filter.DateFrom == nil || scheduledomain.FormatDate(*filter.DateFrom) != "2026-04-15" {
		t.Fatalf("unexpected date_from: %+v", filter.DateFrom)
	}

	if filter.DateTo == nil || scheduledomain.FormatDate(*filter.DateTo) != "2026-05-15" {
		t.Fatalf("unexpected date_to: %+v", filter.DateTo)
	}

	if filter.ScheduleID == nil || *filter.ScheduleID != 42 {
		t.Fatalf("unexpected schedule_id: %+v", filter.ScheduleID)
	}
}

func TestDecodeTaskListInputRejectsInvalidScheduleID(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("GET", "/api/v1/tasks?schedule_id=0", nil)

	_, err := decodeTaskListInput(request)
	if err == nil {
		t.Fatalf("expected invalid schedule_id error")
	}
}
