// Package common
package common

import "time"

const DateLayout = "2006-01-02"

type DateType struct {
	time.Time
}

func (d *DateType) UnmarshalJSON(b []byte) error {
	s := string(b)

	t, err := time.Parse(`"`+DateLayout+`"`, s)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

func (d DateType) MarshalJSON() ([]byte, error) {
	return []byte(d.Format(`"` + DateLayout + `"`)), nil
}

func NewDate(t time.Time) DateType {
	y, m, d := t.Date()

	return DateType{
		Time: time.Date(y, m, d, 0, 0, 0, 0, t.Location()),
	}
}
