package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

func TestDecodeScheduleMutationRequestSpecificDates(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("POST", "/api/v1/task-schedules", strings.NewReader(`{
		"title":"Reports",
		"description":"Specific dates",
		"start_date":"2026-04-15",
		"end_date":"2026-05-15",
		"recurrence":{
			"type":"specific_dates",
			"dates":["2026-04-20","2026-05-01"]
		}
	}`))

	input, err := decodeScheduleMutationRequest(request)
	if err != nil {
		t.Fatalf("decode schedule mutation request: %v", err)
	}

	rule, ok := input.Rule.(scheduledomain.SpecificDatesRule)
	if !ok {
		t.Fatalf("unexpected rule type: %T", input.Rule)
	}

	if input.Title != "Reports" {
		t.Fatalf("unexpected title: got %q", input.Title)
	}

	if len(rule.Dates) != 2 {
		t.Fatalf("unexpected specific dates count: got %d want 2", len(rule.Dates))
	}

	if got := scheduledomain.FormatDate(rule.Dates[0]); got != "2026-04-20" {
		t.Fatalf("unexpected first specific date: got %s", got)
	}
}

func TestRecurrenceDTOToDomainRejectsWeekly(t *testing.T) {
	t.Parallel()

	_, err := recurrenceDTOToDomain(recurrenceDTO{Type: "weekly"})
	if err == nil {
		t.Fatalf("expected weekly recurrence to be rejected")
	}
}
