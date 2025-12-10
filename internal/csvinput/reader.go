package csvinput

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"ganttgen/internal/calendar"
	"ganttgen/internal/model"
)

var (
	requiredColumns = []string{"name", "start", "end", "duration", "depends_on"}
	dateLayout      = "2006-01-02"
)

// Read parses the CSV file and returns tasks with their raw attributes.
func Read(path string) ([]model.Task, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	colIndex, err := mapColumns(header)
	if err != nil {
		return nil, err
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
				return nil, fmt.Errorf("row %d: inconsistent field count", row)
			}
			return nil, fmt.Errorf("row %d: %w", row, err)
		}

		task, err := parseRecord(record, colIndex, row)
		if err != nil {
			return nil, err
		}
		if _, exists := nameSet[task.Name]; exists {
			return nil, fmt.Errorf("row %d: duplicate task name %q", row, task.Name)
		}
		nameSet[task.Name] = struct{}{}
		tasks = append(tasks, task)
		row++
	}

	if err := validateDependencies(tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func mapColumns(header []string) (map[string]int, error) {
	mapped := make(map[string]int)
	for idx, col := range header {
		key := strings.ToLower(strings.TrimSpace(col))
		mapped[key] = idx
	}
	for _, col := range requiredColumns {
		if _, ok := mapped[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}
	return mapped, nil
}

func parseRecord(record []string, col map[string]int, row int) (model.Task, error) {
	get := func(key string) string {
		if idx, ok := col[key]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	name := get("name")
	startStr := get("start")
	endStr := get("end")
	durationStr := get("duration")
	dependsStr := get("depends_on")
	actualStartStr := get("actual_start")
	actualEndStr := get("actual_end")
	actualDurationStr := get("actual_duration")

	if name == "" && startStr == "" && endStr == "" && durationStr == "" && dependsStr == "" && actualStartStr == "" && actualEndStr == "" && actualDurationStr == "" {
		return model.Task{}, fmt.Errorf("row %d: all fields are empty", row)
	}
	if name == "" {
		return model.Task{}, fmt.Errorf("row %d: name is required", row)
	}

	task := model.Task{
		Name:      name,
		DependsOn: parseDepends(dependsStr),
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
	parsed, err := time.Parse(dateLayout, raw)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local), nil
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
