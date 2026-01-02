package calendar

import (
	"sync"
	"time"
)

var (
	holidaysMu sync.RWMutex
	holidays   = map[time.Time]struct{}{}
	// allWorkdays treats weekends and holidays as workdays when true.
	allWorkdays bool
)

// SetHolidays registers dates that should be treated as non-workdays.
// Passing nil clears any previously configured holidays.
func SetHolidays(dates []time.Time) {
	holidaysMu.Lock()
	defer holidaysMu.Unlock()

	holidays = make(map[time.Time]struct{}, len(dates))
	for _, d := range dates {
		holidays[DateOnly(d)] = struct{}{}
	}
}

// SetAllWorkdays controls whether weekends and holidays are treated as workdays.
func SetAllWorkdays(enabled bool) {
	holidaysMu.Lock()
	defer holidaysMu.Unlock()
	allWorkdays = enabled
}

func isHoliday(t time.Time) bool {
	holidaysMu.RLock()
	defer holidaysMu.RUnlock()
	if allWorkdays {
		return false
	}
	_, ok := holidays[DateOnly(t)]
	return ok
}

// IsWorkday reports whether the given date falls on a weekday (Mon-Fri).
func IsWorkday(t time.Time) bool {
	day := DateOnly(t)
	if allWorkdays {
		return true
	}
	if isHoliday(day) {
		return false
	}
	switch day.Weekday() {
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
