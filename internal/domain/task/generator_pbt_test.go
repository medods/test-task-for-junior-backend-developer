package task

import (
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestGeneratorProperties(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 200 
	properties := gopter.NewProperties(parameters)

	genDailyConfig := gen.Struct(reflect.TypeOf(RecurrenceConfig{}), map[string]gopter.Gen{
		"Type":     gen.Const(RecurrenceTypeDaily),
		"Interval": gen.IntRange(1, 30),
		"StartDate": gen.TimeRange(
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC),
		),
		"EndDate":  gen.Maybe(gen.Time()), 
		"Timezone": gen.Const("UTC"),
	})

	properties.Property("daily: dates >= start & unique", prop.ForAll(
		func(cfg *RecurrenceConfig) bool {
			if cfg.EndDate != nil && cfg.EndDate.Before(cfg.StartDate) {
				cfg.EndDate = nil
			}
			dates, err := (&DailyGenerator{}).Generate(*cfg)
			if err != nil {
				return true
			}
			for _, d := range dates {
				if d.Before(cfg.StartDate) {
					return false
				}
			}
			return len(uniqueDates(dates)) == len(dates)
		},
		gopter.Gen{Type: genDailyConfig},
	))

	properties.Property("daily: exact intervals", prop.ForAll(
		func(cfg *RecurrenceConfig) bool {
			cfg.EndDate = nil 
			dates, _ := (&DailyGenerator{}).Generate(*cfg)
			if len(dates) < 2 {
				return true
			}
			interval := time.Duration(*cfg.Interval) * 24 * time.Hour
			for i := 1; i < len(dates); i++ {
				diff := dates[i].Sub(dates[i-1])
				if diff != interval {
					return false 
				}
			}
			return true
		},
		gopter.Gen{Type: genDailyConfig},
	))

	properties.TestingRun(t)
}

func uniqueDates(dates []time.Time) []time.Time {
	seen := make(map[string]struct{}, len(dates))
	var res []time.Time
	for _, d := range dates {
		key := d.UTC().Format("2006-01-02")
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			res = append(res, d)
		}
	}
	return res
}