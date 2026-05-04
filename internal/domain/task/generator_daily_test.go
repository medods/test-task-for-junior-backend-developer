package task

import (
	"testing"
	"time"
)

func TestDailyGenerator_Generate(t *testing.T) {
	t.Parallel()

	utc := time.UTC
	base := time.Date(2026, 5, 1, 10, 0, 0, 0, utc)
	end := time.Date(2026, 5, 10, 0, 0, 0, 0, utc)

	tests := []struct {
		name      string
		cfg       RecurrenceConfig
		wantCount int
		wantErr   bool
	}{
		{
			name: "every 2 days within range",
			cfg: RecurrenceConfig{
				Type:      RecurrenceTypeDaily,
				Interval:  ptrInt(2),
				StartDate: base,
				EndDate:   &end,
			},
			wantCount: 5, 
		},
		{
			name: "interval < 1 returns error",
			cfg: RecurrenceConfig{
				Type:      RecurrenceTypeDaily,
				Interval:  ptrInt(0),
				StartDate: base,
			},
			wantErr: true,
		},
		{
			name: "no end date with limit",
			cfg: RecurrenceConfig{
				Type:      RecurrenceTypeDaily,
				Interval:  ptrInt(1),
				StartDate: base,
			},
		},
	}

	for _, tt := range tests {
		tt := tt 
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := &DailyGenerator{}
			dates, err := g.Generate(tt.cfg)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if tt.wantCount > 0 && len(dates) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(dates), tt.wantCount)
			}

			for i, d := range dates {
				if !d.Equal(d.UTC().Truncate(24 * time.Hour)) {
					t.Errorf("date[%d] not normalized to UTC 00:00:00: %v", i, d)
				}
			}
		})
	}
}

func ptrInt(v int) *int { return &v }