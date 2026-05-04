package task

import (
	"testing"
	"time"
)

func TestRecurrenceConfig_Validate(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	future := now.AddDate(0, 0, 5)
	end := now.AddDate(0, 0, 10)

	tests := []struct {
		name    string
		cfg     RecurrenceConfig
		wantErr error
	}{
		{
			name:    "empty type",
			cfg:     RecurrenceConfig{StartDate: future},
			wantErr: ErrTypeRequired,
		},
		{
			name:    "zero start date",
			cfg:     RecurrenceConfig{Type: RecurrenceTypeDaily},
			wantErr: ErrStartDateRequired,
		},
		{
			name:    "end before start",
			cfg:     RecurrenceConfig{Type: RecurrenceTypeDaily, StartDate: end, EndDate: &now},
			wantErr: ErrStartDateAfterEnd,
		},

		{
			name: "daily valid",
			cfg: RecurrenceConfig{
				Type:     RecurrenceTypeDaily,
				Interval: ptrInt(2),
				StartDate: future,
			},
			wantErr: nil,
		},
		{
			name: "daily interval zero",
			cfg: RecurrenceConfig{
				Type:     RecurrenceTypeDaily,
				Interval: ptrInt(0),
				StartDate: future,
			},
			wantErr: ErrInvalidInterval,
		},
		{
			name: "daily interval nil",
			cfg: RecurrenceConfig{
				Type:     RecurrenceTypeDaily,
				StartDate: future,
			},
			wantErr: ErrInvalidInterval,
		},

		{
			name: "monthly valid",
			cfg: RecurrenceConfig{
				Type:       RecurrenceTypeMonthly,
				DayOfMonth: ptrInt(15),
				StartDate:  future,
			},
			wantErr: nil,
		},
		{
			name: "monthly day 0",
			cfg: RecurrenceConfig{
				Type:       RecurrenceTypeMonthly,
				DayOfMonth: ptrInt(0),
				StartDate:  future,
			},
			wantErr: ErrInvalidDayOfMonth,
		},
		{
			name: "monthly day 32",
			cfg: RecurrenceConfig{
				Type:       RecurrenceTypeMonthly,
				DayOfMonth: ptrInt(32),
				StartDate:  future,
			},
			wantErr: ErrInvalidDayOfMonth,
		},
		{
			name: "monthly invalid overflow",
			cfg: RecurrenceConfig{
				Type:             RecurrenceTypeMonthly,
				DayOfMonth:       ptrInt(15),
				StartDate:        future,
				OverflowStrategy: "invalid",
			},
			wantErr: ErrInvalidOverflowStrategy,
		},

		{
			name: "dates valid",
			cfg: RecurrenceConfig{
				Type:          RecurrenceTypeSpecificDates,
				SpecificDates: []time.Time{future, end},
				StartDate:     future,
			},
			wantErr: nil,
		},
		{
			name: "dates empty",
			cfg: RecurrenceConfig{
				Type:          RecurrenceTypeSpecificDates,
				SpecificDates: []time.Time{},
				StartDate:     future,
			},
			wantErr: ErrEmptyDates,
		},
		{
			name: "dates duplicate",
			cfg: RecurrenceConfig{
				Type:          RecurrenceTypeSpecificDates,
				SpecificDates: []time.Time{future, future},
				StartDate:     future,
			},
			wantErr: ErrDuplicateDate,
		},
		{
			name: "dates zero time",
			cfg: RecurrenceConfig{
				Type:          RecurrenceTypeSpecificDates,
				SpecificDates: []time.Time{future, {}},
				StartDate:     future,
			},
			wantErr: ErrInvalidDateInList,
		},

		// Parity validation
		{
			name: "parity valid even",
			cfg: RecurrenceConfig{
				Type:       RecurrenceTypeParity,
				ParityMode: ptrParity(ParityEven),
				StartDate:  future,
			},
			wantErr: nil,
		},
		{
			name: "parity valid odd",
			cfg: RecurrenceConfig{
				Type:       RecurrenceTypeParity,
				ParityMode: ptrParity(ParityOdd),
				StartDate:  future,
			},
			wantErr: nil,
		},
		{
			name: "parity invalid mode",
			cfg: RecurrenceConfig{
				Type:       RecurrenceTypeParity,
				ParityMode: ptrParity("both"),
				StartDate:  future,
			},
			wantErr: ErrInvalidParityMode,
		},
		{
			name: "parity nil mode",
			cfg: RecurrenceConfig{
				Type:      RecurrenceTypeParity,
				StartDate: future,
			},
			wantErr: ErrInvalidParityMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsValidDayInMonth(t *testing.T) {
	tests := []struct {
		name      string
		year      int
		month     int
		day       int
		wantValid bool
	}{
		{"Feb 28 leap", 2024, 2, 28, true},
		{"Feb 29 leap", 2024, 2, 29, true},
		{"Feb 29 non-leap", 2023, 2, 29, false},
		{"Feb 30 always invalid", 2024, 2, 30, false},
		{"Apr 30", 2026, 4, 30, true},
		{"Apr 31 invalid", 2026, 4, 31, false},
		{"Dec 31", 2026, 12, 31, true},
		{"Jan 1", 2000, 1, 1, true},
		{"Day 0 invalid", 2026, 5, 0, false},
		{"Day 32 invalid", 2026, 5, 32, false},
		{"Month 0 invalid", 2026, 0, 15, false},
		{"Month 13 invalid", 2026, 13, 15, false},
		{"Century leap", 2000, 2, 29, true},
		{"Century non-leap", 1900, 2, 29, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDayInMonth(tt.year, tt.month, tt.day)
			if got != tt.wantValid {
				t.Errorf("IsValidDayInMonth(%d, %d, %d) = %v, want %v",
					tt.year, tt.month, tt.day, got, tt.wantValid)
			}
		})
	}
}

func TestNormalizeToUTCStartOfDay(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Vladivostok") // UTC+10
	local := time.Date(2026, 1, 15, 14, 30, 45, 123, loc)

	utcNormalized := NormalizeToUTCStartOfDay(local)
	expected := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	if !utcNormalized.Equal(expected) {
		t.Errorf("NormalizeToUTCStartOfDay() = %v, want %v", utcNormalized, expected)
	}
}

func TestCalculateEffectiveStart_Daily(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		cfg      RecurrenceConfig
		expected time.Time
	}{
		{
			name: "start in future",
			cfg: RecurrenceConfig{
				Type:      RecurrenceTypeDaily,
				Interval:  ptrInt(1),
				StartDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			},
			expected: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "start in past, next occurrence",
			cfg: RecurrenceConfig{
				Type:      RecurrenceTypeDaily,
				Interval:  ptrInt(2),
				StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.CalculateEffectiveStart(now)
			if !got.Equal(tt.expected) {
				t.Errorf("CalculateEffectiveStart() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func ptrInt(i int) *int {
	return &i
}

func ptrParity(p ParityMode) *ParityMode {
	return &p
}