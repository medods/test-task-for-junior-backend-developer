package dateutil

import "time"

const Layout = "2006-01-02"

func Normalize(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func Parse(value string) (time.Time, error) {
	parsed, err := time.Parse(Layout, value)
	if err != nil {
		return time.Time{}, err
	}

	return Normalize(parsed), nil
}

func Format(t time.Time) string {
	return Normalize(t).Format(Layout)
}

func Today(now func() time.Time) time.Time {
	return Normalize(now())
}

func Max(a, b time.Time) time.Time {
	if a.After(b) {
		return Normalize(a)
	}

	return Normalize(b)
}

func Min(a, b time.Time) time.Time {
	if a.Before(b) {
		return Normalize(a)
	}

	return Normalize(b)
}

func AddDays(t time.Time, days int) time.Time {
	return Normalize(t.AddDate(0, 0, days))
}

func DaysBetween(from, to time.Time) int {
	from = Normalize(from)
	to = Normalize(to)

	return int(to.Sub(from).Hours() / 24)
}
