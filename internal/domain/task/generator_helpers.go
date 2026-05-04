package task

func calculateGenerationLimit(endDate *time.Time) int {
	if endDate == nil {
		return 365 * 2 
	}
	const maxLimit = 10000
	return maxLimit
}