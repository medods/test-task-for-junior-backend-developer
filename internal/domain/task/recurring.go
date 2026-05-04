package task

import (
	"errors"
	"fmt"
	"time"
)

type ReccurenceType string

const (
	TypeDaily   RecurrenceType = "daily"
	TypeMonthly RecurrenceType = "monthly"
	TypeDates   RecurrenceType = "dates"
	TypeParity  RecurrenceType = "parity"
)

type ParityMode string

const (
	ParityEven ParityMode = "even"
	ParityOdd  ParityMode = "odd"
)

type OverflowStrategy string

const (
	OverflowSkip      OverflowStrategy = "skip"      
	OverflowLastDay   OverflowStrategy = "last_day"  
	OverflowNextMonth OverflowStrategy = "next_month" 
)

type RecurrenceConfig struct {
	Type             RecurrenceType   `json:"type"`
	Interval         int              `json:"interval,omitempty"`          
	DayOfMonth       int              `json:"day_of_month,omitempty"`     
	SpecificDates    []time.Time      `json:"dates,omitempty"`            
	ParityMode       ParityMode       `json:"parity_mode,omitempty"`       
	StartDate        time.Time        `json:"start_date"`
	EndDate          *time.Time       `json:"end_date,omitempty"`
	Timezone         string           `json:"timezone,omitempty"`
	OverflowStrategy OverflowStrategy `json:"overflow_strategy,omitempty"` 
}
