package calendar

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, y int, m time.Month, d int) time.Time {
	t.Helper()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func TestIsWorkday(t *testing.T) {
	if !IsWorkday(mustDate(t, 2024, time.June, 3)) { // Monday
		t.Fatalf("expected Monday to be workday")
	}
	if IsWorkday(mustDate(t, 2024, time.June, 2)) { // Sunday
		t.Fatalf("expected Sunday to be non-workday")
	}
}

func TestNextWorkday(t *testing.T) {
	sat := mustDate(t, 2024, time.June, 1) // Saturday
	if got := NextWorkday(sat); got.Weekday() != time.Monday {
		t.Fatalf("expected Monday, got %v", got.Weekday())
	}
	mon := mustDate(t, 2024, time.June, 3)
	if got := NextWorkday(mon); !got.Equal(mon) {
		t.Fatalf("expected same day for workday input")
	}
}

func TestAddWorkdays(t *testing.T) {
	start := mustDate(t, 2024, time.May, 31) // Friday
	got := AddWorkdays(start, 1)             // 1 workday forward -> Monday
	want := mustDate(t, 2024, time.June, 3)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
