package csvinput

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"ganttgen/internal/calendar"
	"ganttgen/internal/model"
)

var (
	requiredColumns = []string{"name", "start", "end", "duration", "depends_on"}
	columnAliases   = map[string]string{
		"タスク名":   "name",
		"開始":     "start",
		"終了":     "end",
		"期間":     "duration",
		"依存":     "depends_on",
		"実績開始":   "actual_start",
		"実績終了":   "actual_end",
		"実績期間":   "actual_duration",
		"状態":     "status",
		"備考":     "notes",
		"notes":  "notes",
		"status": "status",
	}
	knownColumns = map[string]struct{}{
		"name":            {},
		"start":           {},
		"end":             {},
		"duration":        {},
		"depends_on":      {},
		"actual_start":    {},
		"actual_end":      {},
		"actual_duration": {},
		"status":          {},
		"notes":           {},
	}
	dateLayouts = []string{
		"2006-01-02", // zero-padded dash
		"2006-1-2",   // non-padded dash
		"2006/01/02", // zero-padded slash
		"2006/1/2",   // non-padded slash
	}
)

type customColumn struct {
	Name  string
	Index int
}

// Read parses the CSV file and returns tasks with their raw attributes.
func Read(path string) ([]model.Task, []string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open csv: %w", err)
	}

	decoded, err := decodeCSVBytes(raw)
	if err != nil {
		return nil, nil, err
	}

	reader := csv.NewReader(bytes.NewReader(decoded))
	header, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("read header: %w", err)
	}

	colIndex, customCols, err := mapColumns(header)
	if err != nil {
		return nil, nil, err
	}

	var tasks []model.Task
	nameSet := make(map[string]struct{})
	row := 2 // 1-based row number, header is 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				return nil, nil, fmt.Errorf("row %d: inconsistent field count", row)
			}
			return nil, nil, fmt.Errorf("row %d: %w", row, err)
		}

		if recordAllEmpty(record) {
			row++
			continue
		}

		task, err := parseRecord(record, colIndex, customCols, row)
		if err != nil {
			return nil, nil, err
		}
		if task.IsHeading {
			tasks = append(tasks, task)
			row++
			continue
		}
		if _, exists := nameSet[task.Name]; exists {
			return nil, nil, fmt.Errorf("row %d: duplicate task name %q", row, task.Name)
		}
		nameSet[task.Name] = struct{}{}
		tasks = append(tasks, task)
		row++
	}

	if err := validateDependencies(tasks); err != nil {
		return nil, nil, err
	}

	return tasks, customColumnNames(customCols), nil
}

func mapColumns(header []string) (map[string]int, []customColumn, error) {
	mapped := make(map[string]int)
	var customCols []customColumn
	seenCustom := make(map[string]struct{})
	for idx, col := range header {
		key := strings.ToLower(strings.TrimSpace(col))
		if canonical, ok := columnAliases[key]; ok {
			key = canonical
		}
		mapped[key] = idx
		if _, ok := knownColumns[key]; !ok {
			trimmed := strings.TrimSpace(col)
			if trimmed != "" {
				if _, seen := seenCustom[trimmed]; !seen {
					seenCustom[trimmed] = struct{}{}
					customCols = append(customCols, customColumn{Name: trimmed, Index: idx})
				}
			}
		}
	}
	for _, col := range requiredColumns {
		if _, ok := mapped[col]; !ok {
			return nil, nil, fmt.Errorf("missing required column: %s", col)
		}
	}
	return mapped, customCols, nil
}

func parseRecord(record []string, col map[string]int, customCols []customColumn, row int) (model.Task, error) {
	get := func(key string) string {
		if idx, ok := col[key]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	customValues := makeCustomValues(record, customCols)
	name := get("name")
	statusStr := get("status")
	if strings.HasPrefix(name, "#") {
		return model.Task{
			Name:         strings.TrimSpace(strings.TrimPrefix(name, "#")),
			IsHeading:    true,
			Status:       statusStr,
			Notes:        get("notes"),
			CustomValues: customValues,
		}, nil
	}
	startStr := get("start")
	endStr := get("end")
	durationStr := get("duration")
	dependsStr := get("depends_on")
	actualStartStr := get("actual_start")
	actualEndStr := get("actual_end")
	actualDurationStr := get("actual_duration")
	notesStr := get("notes")

	// Name only (no scheduling/depends/actual) -> display-only row (notes allowed).
	if name != "" && startStr == "" && endStr == "" && durationStr == "" && dependsStr == "" && actualStartStr == "" && actualEndStr == "" && actualDurationStr == "" {
		return model.Task{Name: name, DisplayOnly: true, Notes: notesStr, CustomValues: customValues}, nil
	}

	if name == "" {
		return model.Task{}, fmt.Errorf("row %d: name is required", row)
	}

	task := model.Task{
		Name:         name,
		DependsOn:    parseDepends(dependsStr),
		Notes:        notesStr,
		Status:       statusStr,
		CustomValues: customValues,
	}

	if startStr != "" {
		parsed, err := parseDate(startStr)
		if err != nil {
			return model.Task{}, fmt.Errorf("row %d: invalid start: %w", row, err)
		}
		task.Start = &parsed
	}

	if endStr != "" {
		parsed, err := parseDate(endStr)
		if err != nil {
			return model.Task{}, fmt.Errorf("row %d: invalid end: %w", row, err)
		}
		task.End = &parsed
	}

	if durationStr != "" {
		days, err := parseDuration(durationStr)
		if err != nil {
			return model.Task{}, fmt.Errorf("row %d: invalid duration: %w", row, err)
		}
		task.DurationDays = days
	}

	if task.End != nil && task.DurationDays > 0 {
		return model.Task{}, fmt.Errorf("row %d: end and duration cannot both be set", row)
	}
	if task.End != nil && task.Start == nil && task.DurationDays == 0 {
		return model.Task{}, fmt.Errorf("row %d: end cannot be set without start or duration", row)
	}
	if task.DurationDays == 0 && task.End == nil {
		return model.Task{}, fmt.Errorf("row %d: either duration or end must be provided", row)
	}
	if task.Start == nil && task.DurationDays > 0 && len(task.DependsOn) == 0 {
		return model.Task{}, fmt.Errorf("row %d: duration-only task must depend on another task or define a start", row)
	}
	if task.Start == nil && task.End == nil && task.DurationDays == 0 {
		return model.Task{}, fmt.Errorf("row %d: task lacks scheduling information", row)
	}

	if err := parseActual(&task, actualStartStr, actualEndStr, actualDurationStr, row); err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func makeCustomValues(record []string, customCols []customColumn) []string {
	if len(customCols) == 0 {
		return nil
	}
	values := make([]string, len(customCols))
	for i, col := range customCols {
		if col.Index < len(record) {
			values[i] = strings.TrimSpace(record[col.Index])
		}
	}
	return values
}

func customColumnNames(customCols []customColumn) []string {
	if len(customCols) == 0 {
		return nil
	}
	names := make([]string, len(customCols))
	for i, col := range customCols {
		names[i] = col.Name
	}
	return names
}

func parseDepends(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})

	var deps []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			deps = append(deps, trimmed)
		}
	}
	return deps
}

func parseDate(raw string) (time.Time, error) {
	for _, layout := range dateLayouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q (expected YYYY-MM-DD or YYYY/MM/DD)", raw)
}

func parseDuration(raw string) (int, error) {
	if len(raw) < 2 {
		return 0, errors.New("duration must be Nd (e.g. 5d)")
	}
	if raw[len(raw)-1] != 'd' && raw[len(raw)-1] != 'D' {
		return 0, errors.New("duration must end with 'd'")
	}
	num := raw[:len(raw)-1]
	days, err := strconv.Atoi(num)
	if err != nil || days <= 0 {
		return 0, errors.New("duration must be a positive integer followed by 'd'")
	}
	return days, nil
}

func validateDependencies(tasks []model.Task) error {
	nameSet := make(map[string]struct{}, len(tasks))
	for _, t := range tasks {
		nameSet[t.Name] = struct{}{}
	}

	for _, t := range tasks {
		for _, dep := range t.DependsOn {
			if _, ok := nameSet[dep]; !ok {
				return fmt.Errorf("task %q depends on unknown task %q", t.Name, dep)
			}
		}
	}
	return nil
}

func parseActual(task *model.Task, startStr, endStr, durationStr string, row int) error {
	if startStr == "" && endStr == "" && durationStr == "" {
		return nil
	}

	if startStr != "" {
		parsed, err := parseDate(startStr)
		if err != nil {
			return fmt.Errorf("row %d: invalid actual_start: %w", row, err)
		}
		task.ActualStart = ptrTime(calendar.NextWorkday(parsed))
	}
	if endStr != "" {
		parsed, err := parseDate(endStr)
		if err != nil {
			return fmt.Errorf("row %d: invalid actual_end: %w", row, err)
		}
		task.ActualEnd = ptrTime(calendar.NextWorkday(parsed))
	}
	if durationStr != "" {
		days, err := parseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("row %d: invalid actual_duration: %w", row, err)
		}
		task.ActualDurationDays = days
	}

	if task.ActualEnd != nil && task.ActualDurationDays > 0 {
		return fmt.Errorf("row %d: actual_end and actual_duration cannot both be set", row)
	}
	if task.ActualEnd != nil && task.ActualStart == nil && task.ActualDurationDays == 0 {
		return fmt.Errorf("row %d: actual_end cannot be set without actual_start or actual_duration", row)
	}
	if task.ActualDurationDays > 0 && task.ActualStart == nil && task.ActualEnd == nil {
		return fmt.Errorf("row %d: actual_duration requires actual_start", row)
	}

	switch {
	case task.ActualStart != nil && task.ActualEnd != nil:
		if task.ActualEnd.Before(*task.ActualStart) {
			return fmt.Errorf("row %d: actual_end is before actual_start", row)
		}
		task.ComputedActualStart = ptrTime(calendar.DateOnly(*task.ActualStart))
		task.ComputedActualEnd = ptrTime(calendar.DateOnly(*task.ActualEnd))
	case task.ActualStart != nil && task.ActualDurationDays > 0:
		start := calendar.DateOnly(*task.ActualStart)
		end := calendar.AddWorkdays(start, task.ActualDurationDays-1)
		task.ComputedActualStart = &start
		task.ComputedActualEnd = &end
	case task.ActualStart != nil:
		start := calendar.DateOnly(*task.ActualStart)
		task.ComputedActualStart = &start
		task.ComputedActualEnd = &start
	default:
		return fmt.Errorf("row %d: actual schedule cannot be determined", row)
	}

	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func recordAllEmpty(record []string) bool {
	for _, v := range record {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

type encodingKind int

const (
	encUTF8 encodingKind = iota
	encShiftJIS
)

// decodeCSVBytes inspects the header to decide UTF-8 or Shift_JIS and returns UTF-8 bytes.
func decodeCSVBytes(data []byte) ([]byte, error) {
	enc := detectEncoding(data)
	switch enc {
	case encUTF8:
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("csv is not valid utf-8")
		}
		return data, nil
	case encShiftJIS:
		reader := transform.NewReader(bytes.NewReader(data), japanese.ShiftJIS.NewDecoder())
		decoded, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("decode Shift_JIS: %w", err)
		}
		return decoded, nil
	default:
		return nil, fmt.Errorf("unknown encoding")
	}
}

// detectEncoding uses the first line (header) as a sample and falls back to Shift_JIS if not valid UTF-8.
func detectEncoding(data []byte) encodingKind {
	sample := data
	if idx := bytes.IndexByte(data, '\n'); idx >= 0 {
		sample = data[:idx]
	}
	if utf8.Valid(sample) {
		return encUTF8
	}
	return encShiftJIS
}
