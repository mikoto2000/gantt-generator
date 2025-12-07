package calendar

import (
	"os"
	"testing"
	"time"
)

func mustDate(t *testing.T, y int, m time.Month, d int) time.Time {
	t.Helper()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func TestIsWorkday(t *testing.T) {
	t.Cleanup(func() { SetHolidays(nil) })

	if !IsWorkday(mustDate(t, 2024, time.June, 3)) { // Monday
		t.Fatalf("expected Monday to be workday")
	}
	if IsWorkday(mustDate(t, 2024, time.June, 2)) { // Sunday
		t.Fatalf("expected Sunday to be non-workday")
	}
}

func TestIsWorkday_Holiday(t *testing.T) {
	SetHolidays([]time.Time{mustDate(t, 2024, time.July, 15)}) // Monday but holiday
	t.Cleanup(func() { SetHolidays(nil) })

	if IsWorkday(mustDate(t, 2024, time.July, 15)) {
		t.Fatalf("expected configured holiday to be non-workday")
	}
	if !IsWorkday(mustDate(t, 2024, time.July, 16)) {
		t.Fatalf("expected next weekday to remain workday")
	}
}

func TestNextWorkday(t *testing.T) {
	t.Cleanup(func() { SetHolidays(nil) })

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
	t.Cleanup(func() { SetHolidays(nil) })

	start := mustDate(t, 2024, time.May, 31) // Friday
	got := AddWorkdays(start, 1)             // 1 workday forward -> Monday
	want := mustDate(t, 2024, time.June, 3)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadHolidaysYAML(t *testing.T) {
	t.Cleanup(func() { SetHolidays(nil) })

	dir := t.TempDir()
	path := dir + "/holidays.yaml"
	content := `
holidays:
  - 2024-09-16
  - 2024-09-23
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}

	if err := LoadHolidaysYAML(path); err != nil {
		t.Fatalf("load holidays: %v", err)
	}

	if IsWorkday(mustDate(t, 2024, time.September, 16)) {
		t.Fatalf("expected holiday from YAML to be non-workday")
	}
	if !IsWorkday(mustDate(t, 2024, time.September, 17)) {
		t.Fatalf("expected adjacent weekday to be workday")
	}
}
