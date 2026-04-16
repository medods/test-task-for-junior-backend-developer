// Package common contains shared domain helpers.
package common

import "time"

// DateLayout is the YYYY-MM-DD layout used for JSON date values.
const DateLayout = "2006-01-02"

// DateType is a date-only value serialized as YYYY-MM-DD.
type DateType struct {
	time.Time
}

// UnmarshalJSON parses a JSON string in YYYY-MM-DD format.
func (d *DateType) UnmarshalJSON(b []byte) error {
	s := string(b)

	t, err := time.Parse(`"`+DateLayout+`"`, s)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

// MarshalJSON formats the date as a JSON string in YYYY-MM-DD format.
func (d DateType) MarshalJSON() ([]byte, error) {
	return []byte(d.Format(`"` + DateLayout + `"`)), nil
}

// NewDate normalizes a time to a date-only value at midnight.
func NewDate(t time.Time) DateType {
	y, m, d := t.Date()

	return DateType{
		Time: time.Date(y, m, d, 0, 0, 0, 0, t.Location()),
	}
}
