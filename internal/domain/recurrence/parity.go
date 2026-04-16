package recurrence

// Parity represents even/odd parity used by MonthParity recurrence.
type Parity string

const (
	NoneParity Parity = "none"
	Even       Parity = "even"
	Odd        Parity = "odd"
)
