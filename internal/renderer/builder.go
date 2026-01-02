package renderer

import (
	"errors"
	"html/template"
	"strconv"
	"strings"
	"time"

	"ganttgen/internal/calendar"
	"ganttgen/internal/model"
)

// BuildHTML prepares render data and returns the final HTML string.
// liveReloadURL, when non-empty, injects a small client to auto-refresh the page.
func BuildHTML(tasks []model.Task, liveReloadURL string, customColumns []string) (string, error) {
	if len(tasks) == 0 {
		return "", errors.New("no tasks to render")
	}

	var (
		minStart time.Time
		maxEnd   time.Time
		setRange bool
	)
	for _, t := range tasks {
		if t.IsHeading || t.DisplayOnly {
			continue
		}
		if !setRange {
			minStart, maxEnd = t.ComputedStart, t.ComputedEnd
			setRange = true
		} else {
			if t.ComputedStart.Before(minStart) {
				minStart = t.ComputedStart
			}
			if t.ComputedEnd.After(maxEnd) {
				maxEnd = t.ComputedEnd
			}
		}
	}
	if !setRange {
		return "", errors.New("no schedulable tasks to render")
	}
	for _, t := range tasks {
		if t.IsHeading || t.DisplayOnly {
			continue
		}
		if t.HasActual() {
			if t.ComputedActualStart.Before(minStart) {
				minStart = *t.ComputedActualStart
			}
			if t.ComputedActualEnd.After(maxEnd) {
				maxEnd = *t.ComputedActualEnd
			}
		}
	}

	today := calendar.DateOnly(time.Now())
	if today.Before(minStart) {
		minStart = today
	}
	if today.After(maxEnd) {
		maxEnd = today
	}

	days := daysRange(minStart, maxEnd)
	todayIndex := daysBetween(minStart, today)

	var rows []renderRow
	var hasActual bool
	var hasNotes bool
	customCount := len(customColumns)
	for _, t := range tasks {
		customValues := padCustomValues(t.CustomValues, customCount)
		if t.IsHeading {
			if t.Notes != "" {
				hasNotes = true
			}
			rows = append(rows, renderRow{
				Heading:       t.Name,
				HeadingStatus: t.Status,
				HeadingNotes:  t.Notes,
				CustomValues:  customValues,
				FilterName:    t.Name,
				FilterStatus:  t.Status,
				FilterNotes:   t.Notes,
			})
			continue
		}
		if t.DisplayOnly {
			if t.Notes != "" {
				hasNotes = true
			}
			rows = append(rows, renderRow{
				DisplayOnly:      t.Name,
				DisplayOnlyNotes: t.Notes,
				CustomValues:     customValues,
				FilterName:       t.Name,
				FilterStatus:     "",
				FilterNotes:      t.Notes,
			})
			continue
		}
		startIdx := daysBetween(minStart, t.ComputedStart)
		span := daysBetween(t.ComputedStart, t.ComputedEnd) + 1
		rt := renderTask{
			Name:       t.Name,
			Status:     t.Status,
			Notes:      t.Notes,
			Cancelled:  t.IsCancelled(),
			StartIndex: startIdx,
			Span:       span,
			Start:      calendar.DateOnly(t.ComputedStart),
			End:        calendar.DateOnly(t.ComputedEnd),
		}
		if t.HasActual() {
			hasActual = true
			actualStartIdx := daysBetween(minStart, *t.ComputedActualStart)
			actualSpan := daysBetween(*t.ComputedActualStart, *t.ComputedActualEnd) + 1
			rt.Actual = &renderActual{
				StartIndex: actualStartIdx,
				Span:       actualSpan,
				Start:      calendar.DateOnly(*t.ComputedActualStart),
				End:        calendar.DateOnly(*t.ComputedActualEnd),
			}
		}
		if t.Notes != "" {
			hasNotes = true
		}
		rows = append(rows, renderRow{
			Task:         &rt,
			CustomValues: customValues,
			FilterName:   t.Name,
			FilterStatus: t.Status,
			FilterNotes:  t.Notes,
		})
	}

	filterColumns := buildFilterColumns(rows, hasNotes, customColumns)

	ctx := renderContext{
		Days:              days,
		Rows:              rows,
		DayCount:          len(days),
		TodayIndex:        todayIndex,
		HasActual:         hasActual,
		HasNotes:          hasNotes,
		HasCustomColumns:  customCount > 0,
		CustomColumns:     customColumns,
		CustomColumnCount: customCount,
		FilterColumns:     filterColumns,
		LiveReloadURL:     liveReloadURL,
		CSS:               template.CSS(baseCSS()),
	}
	return renderHTML(ctx)
}

func daysRange(start, end time.Time) []time.Time {
	var res []time.Time
	for d := calendar.DateOnly(start); !d.After(end); d = d.AddDate(0, 0, 1) {
		res = append(res, d)
	}
	return res
}

func daysBetween(start, end time.Time) int {
	days := 0
	for d := calendar.DateOnly(start); d.Before(calendar.DateOnly(end)); d = d.AddDate(0, 0, 1) {
		days++
	}
	return days
}

func padCustomValues(values []string, count int) []string {
	if count == 0 {
		return nil
	}
	if len(values) >= count {
		return values[:count]
	}
	padded := make([]string, count)
	copy(padded, values)
	return padded
}

func buildFilterColumns(rows []renderRow, hasNotes bool, customColumns []string) []filterColumn {
	names := make([]string, 0, len(rows))
	statuses := make([]string, 0, len(rows))
	notes := make([]string, 0, len(rows))
	customValues := make([][]string, len(customColumns))

	for _, row := range rows {
		names = append(names, row.FilterName)
		statuses = append(statuses, row.FilterStatus)
		notes = append(notes, row.FilterNotes)
		for i := range customColumns {
			if i < len(row.CustomValues) {
				customValues[i] = append(customValues[i], row.CustomValues[i])
			} else {
				customValues[i] = append(customValues[i], "")
			}
		}
	}

	filterColumns := []filterColumn{
		{Key: "name", Label: "Task", Values: uniqueValues(names)},
		{Key: "status", Label: "状態", Values: uniqueValues(statuses)},
	}
	if hasNotes {
		filterColumns = append(filterColumns, filterColumn{Key: "notes", Label: "備考", Values: uniqueValues(notes)})
	}
	for i, col := range customColumns {
		filterColumns = append(filterColumns, filterColumn{
			Key:    "custom-" + strconv.Itoa(i),
			Label:  col,
			Values: uniqueValues(customValues[i]),
		})
	}
	return filterColumns
}

func uniqueValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var res []string
	hasEmpty := false
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			hasEmpty = true
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		res = append(res, trimmed)
	}
	if hasEmpty {
		res = append(res, "")
	}
	return res
}

type renderTask struct {
	Name       string
	Status     string
	Notes      string
	Cancelled  bool
	StartIndex int
	Span       int
	Start      time.Time
	End        time.Time
	Actual     *renderActual
}

type renderRow struct {
	Heading          string
	HeadingStatus    string
	HeadingNotes     string
	DisplayOnly      string
	DisplayOnlyNotes string
	Task             *renderTask
	CustomValues     []string
	FilterName       string
	FilterStatus     string
	FilterNotes      string
}

type renderActual struct {
	StartIndex int
	Span       int
	Start      time.Time
	End        time.Time
}

type renderContext struct {
	Days              []time.Time
	Rows              []renderRow
	DayCount          int
	TodayIndex        int
	HasActual         bool
	HasNotes          bool
	HasCustomColumns  bool
	CustomColumns     []string
	CustomColumnCount int
	FilterColumns     []filterColumn
	LiveReloadURL     string
	CSS               template.CSS
}

type filterColumn struct {
	Key    string
	Label  string
	Values []string
}
