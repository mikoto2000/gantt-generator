package calendar

import (
	"bytes"
	"os"
	"testing"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func mustDate(t *testing.T, y int, m time.Month, d int) time.Time {
	t.Helper()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func TestIsWorkday(t *testing.T) {
	t.Cleanup(func() { SetHolidays(nil) })
	t.Cleanup(func() { SetAllWorkdays(false) })

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
	t.Cleanup(func() { SetAllWorkdays(false) })

	if IsWorkday(mustDate(t, 2024, time.July, 15)) {
		t.Fatalf("expected configured holiday to be non-workday")
	}
	if !IsWorkday(mustDate(t, 2024, time.July, 16)) {
		t.Fatalf("expected next weekday to remain workday")
	}
}

func TestIsWorkday_AllWorkdays(t *testing.T) {
	SetHolidays([]time.Time{mustDate(t, 2024, time.July, 15)})
	SetAllWorkdays(true)
	t.Cleanup(func() { SetHolidays(nil) })
	t.Cleanup(func() { SetAllWorkdays(false) })

	if !IsWorkday(mustDate(t, 2024, time.July, 15)) {
		t.Fatalf("expected configured holiday to be workday when overridden")
	}
	if !IsWorkday(mustDate(t, 2024, time.June, 1)) { // Saturday
		t.Fatalf("expected weekend to be workday when overridden")
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

func TestLoadHolidaysCSV(t *testing.T) {
	t.Cleanup(func() { SetHolidays(nil) })

	dir := t.TempDir()
	path := dir + "/syukujitsu.csv"
	var encoded bytes.Buffer
	writer := transform.NewWriter(&encoded, japanese.ShiftJIS.NewEncoder())
	content := "国民の祝日・休日月日,国民の祝日・休日名称\n2024/9/16,敬老の日\n2024/9/23,秋分の日\n"
	if _, err := writer.Write([]byte(content)); err != nil {
		t.Fatalf("encode temp csv: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close temp csv encoder: %v", err)
	}
	if err := os.WriteFile(path, encoded.Bytes(), 0o644); err != nil {
		t.Fatalf("write temp csv: %v", err)
	}

	if err := LoadHolidaysCSV(path); err != nil {
		t.Fatalf("load holidays csv: %v", err)
	}

	if IsWorkday(mustDate(t, 2024, time.September, 16)) {
		t.Fatalf("expected holiday from CSV to be non-workday")
	}
	if !IsWorkday(mustDate(t, 2024, time.September, 17)) {
		t.Fatalf("expected adjacent weekday to be workday")
	}
}
