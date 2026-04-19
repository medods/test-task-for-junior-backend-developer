package task

import (
	"sort"
	"time"
)

// GenerateRecurringDates генерирует даты для повторяющейся задачи
func GenerateRecurringDates(config *RecurrenceConfig, startDate time.Time, now time.Time) ([]time.Time, error) {
	if config == nil {
		return nil, nil
	}

	repeatCount := config.RepeatCount
	if repeatCount == 0 {
		repeatCount = getDefaultRepeatCount(config.Type)
	}

	var dates []time.Time
	var err error

	switch config.Type {
	case "daily":
		dates = generateDailyDates(config.Daily, startDate, repeatCount)
	case "monthly":
		dates = generateMonthlyDates(config.Monthly, startDate, repeatCount)
	case "specific_dates":
		dates, err = generateSpecificDates(config.SpecificDates, startDate)
		if err != nil {
			return nil, err
		}
		repeatCount = min(repeatCount, len(dates))
	case "parity":
		dates = generateParityDates(config.ParityDates, startDate, repeatCount)
	default:
		return nil, nil
	}

	return filterDates(dates, now, repeatCount), nil
}

func filterDates(dates []time.Time, now time.Time, limit int) []time.Time {
	filtered := make([]time.Time, 0, limit)
	for i, d := range dates {
		if i >= limit {
			break
		}
		if !d.Before(now) { // d.After(now) || d.Equal(now)
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func generateDailyDates(intervalDays int, startDate time.Time, count int) []time.Time {
	if count <= 0 {
		return []time.Time{}
	}
	dates := make([]time.Time, 0, count)
	for i := 0; i < count; i++ {
		dates = append(dates, startDate.AddDate(0, 0, i*intervalDays))
	}
	return dates
}

func generateMonthlyDates(monthDay int, startDate time.Time, count int) []time.Time {
	if count <= 0 {
		return []time.Time{}
	}
	dates := make([]time.Time, 0, count)
	year := startDate.Year()
	month := startDate.Month()

	for len(dates) < count {
		targetDate := time.Date(year, month, monthDay, 0, 0, 0, 0, time.UTC)
		//сдвиг даты на крайнюю в месяце
		if targetDate.Month() != month {
			lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
			targetDate = lastDay
		}
		dates = append(dates, targetDate)
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
	return dates
}

func generateSpecificDates(specificDates []string, startDate time.Time) ([]time.Time, error) {
	if len(specificDates) == 0 {
		return []time.Time{}, nil
	}
	//проверка на дубли
	uniqueDates := make(map[string]bool)
	for _, dateStr := range specificDates {
		uniqueDates[dateStr] = true
	}

	dates := make([]time.Time, 0, len(uniqueDates))
	for dateStr := range uniqueDates {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		result := time.Date(date.Year(), date.Month(), date.Day(),
			startDate.Hour(), startDate.Minute(), startDate.Second(), 0, time.UTC)
		dates = append(dates, result)
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})
	return dates, nil
}

func generateParityDates(parityType string, startDate time.Time, count int) []time.Time {
	if count <= 0 {
		return []time.Time{}
	}
	dates := make([]time.Time, 0, count)
	currentDate := startDate

	for len(dates) < count {
		day := currentDate.Day()
		isEven := day%2 == 0

		if (parityType == "even" && isEven) || (parityType == "odd" && !isEven) {
			dates = append(dates, currentDate)
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	return dates
}
