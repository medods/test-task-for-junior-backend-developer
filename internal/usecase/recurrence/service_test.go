package recurrence

import (
	"context"
	"testing"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	"example.com/taskservice/internal/shared/dateutil"
)

type stubRepository struct {
	createdRule *recurrencedomain.Rule
	syncedRule  *recurrencedomain.Rule
	syncedDates []time.Time
	listActive  []recurrencedomain.Rule
}

func (s *stubRepository) Create(_ context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error) {
	copied := *rule
	copied.ID = 1
	s.createdRule = &copied
	return &copied, nil
}

func (s *stubRepository) GetByID(_ context.Context, id int64) (*recurrencedomain.Rule, error) {
	if s.createdRule != nil && s.createdRule.ID == id {
		copied := *s.createdRule
		return &copied, nil
	}
	return nil, recurrencedomain.ErrNotFound
}

func (s *stubRepository) Update(_ context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error) {
	copied := *rule
	s.createdRule = &copied
	return &copied, nil
}

func (s *stubRepository) Delete(_ context.Context, _ int64, _ time.Time) error    { return nil }
func (s *stubRepository) List(_ context.Context) ([]recurrencedomain.Rule, error) { return nil, nil }
func (s *stubRepository) ListActive(_ context.Context) ([]recurrencedomain.Rule, error) {
	return s.listActive, nil
}
func (s *stubRepository) SyncGeneratedTasks(_ context.Context, rule *recurrencedomain.Rule, _ time.Time, occurrences []time.Time) error {
	copied := *rule
	s.syncedRule = &copied
	s.syncedDates = append([]time.Time(nil), occurrences...)
	return nil
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := dateutil.Parse(value)
	if err != nil {
		t.Fatalf("parse date %s: %v", value, err)
	}
	return parsed
}

func TestCreateMaterializesRule(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	service.now = func() time.Time { return mustDate(t, "2026-04-20") }
	service.horizonInDays = 4

	interval := 2
	created, err := service.Create(context.Background(), CreateInput{
		Title:        "Inventory",
		ScheduleType: recurrencedomain.ScheduleTypeEveryNDays,
		StartsOn:     ptrTime(mustDate(t, "2026-04-20")),
		EveryNDays:   &interval,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if created.ID != 1 {
		t.Fatalf("expected created rule id 1, got %d", created.ID)
	}
	if repo.syncedRule == nil {
		t.Fatal("expected sync to be called")
	}

	expected := []string{"2026-04-20", "2026-04-22", "2026-04-24"}
	if len(repo.syncedDates) != len(expected) {
		t.Fatalf("expected %d generated dates, got %d", len(expected), len(repo.syncedDates))
	}
	for i, date := range repo.syncedDates {
		if got := dateutil.Format(date); got != expected[i] {
			t.Fatalf("generated date %d: expected %s, got %s", i, expected[i], got)
		}
	}
}

func TestValidateSpecificDatesDeduplicates(t *testing.T) {
	today := mustDate(t, "2026-04-20")
	normalized, err := validateInput(CreateInput{
		Title:         "Call patients",
		ScheduleType:  recurrencedomain.ScheduleTypeSpecificDates,
		SpecificDates: []time.Time{mustDate(t, "2026-04-22"), mustDate(t, "2026-04-22"), mustDate(t, "2026-04-25")},
	}, today)
	if err != nil {
		t.Fatalf("validateInput returned error: %v", err)
	}

	if len(normalized.SpecificDates) != 2 {
		t.Fatalf("expected 2 dates after deduplication, got %d", len(normalized.SpecificDates))
	}
	if got := dateutil.Format(normalized.StartsOn); got != "2026-04-22" {
		t.Fatalf("expected starts_on 2026-04-22, got %s", got)
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
