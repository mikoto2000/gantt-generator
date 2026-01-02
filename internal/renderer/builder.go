package renderer

import (
	"errors"
	"html/template"
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
			rows = append(rows, renderRow{Heading: t.Name, HeadingStatus: t.Status, HeadingNotes: t.Notes, CustomValues: customValues})
			continue
		}
		if t.DisplayOnly {
			if t.Notes != "" {
				hasNotes = true
			}
			rows = append(rows, renderRow{DisplayOnly: t.Name, DisplayOnlyNotes: t.Notes, CustomValues: customValues})
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
		rows = append(rows, renderRow{Task: &rt, CustomValues: customValues})
	}

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
	LiveReloadURL     string
	CSS               template.CSS
}
