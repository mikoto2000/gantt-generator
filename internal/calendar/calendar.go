package calendar

import "time"

// IsWorkday reports whether the given date falls on a weekday (Mon-Fri).
func IsWorkday(t time.Time) bool {
	switch t.Weekday() {
	case time.Saturday, time.Sunday:
		return false
	default:
		return true
	}
}

// DateOnly drops the time component for consistent date math.
func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// NextWorkday returns the same date if it is a workday, or the next workday otherwise.
func NextWorkday(t time.Time) time.Time {
	day := DateOnly(t)
	for !IsWorkday(day) {
		day = day.AddDate(0, 0, 1)
	}
	return day
}

// NextWorkdayAfter returns the next workday strictly after the provided date.
func NextWorkdayAfter(t time.Time) time.Time {
	return NextWorkday(DateOnly(t).AddDate(0, 0, 1))
}

// AddWorkdays moves forward by the given number of workdays (0 keeps the same day).
func AddWorkdays(start time.Time, days int) time.Time {
	current := NextWorkday(start)
	for i := 0; i < days; i++ {
		current = NextWorkday(current.AddDate(0, 0, 1))
	}
	return current
}
