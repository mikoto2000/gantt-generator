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
func BuildHTML(tasks []model.Task, liveReloadURL string) (string, error) {
	if len(tasks) == 0 {
		return "", errors.New("no tasks to render")
	}

	var (
		minStart time.Time
		maxEnd   time.Time
		setRange bool
	)
	for _, t := range tasks {
		if t.IsHeading {
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
		if t.IsHeading {
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
	for _, t := range tasks {
		if t.IsHeading {
			rows = append(rows, renderRow{Heading: t.Name})
			continue
		}
		startIdx := daysBetween(minStart, t.ComputedStart)
		span := daysBetween(t.ComputedStart, t.ComputedEnd) + 1
		rt := renderTask{
			Name:       t.Name,
			Notes:      t.Notes,
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
		rows = append(rows, renderRow{Task: &rt})
	}

	ctx := renderContext{
		Days:          days,
		Rows:          rows,
		DayCount:      len(days),
		TodayIndex:    todayIndex,
		HasActual:     hasActual,
		HasNotes:      hasNotes,
		LiveReloadURL: liveReloadURL,
		CSS:           template.CSS(baseCSS()),
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

type renderTask struct {
	Name       string
	Notes      string
	StartIndex int
	Span       int
	Start      time.Time
	End        time.Time
	Actual     *renderActual
}

type renderRow struct {
	Heading string
	Task    *renderTask
}

type renderActual struct {
	StartIndex int
	Span       int
	Start      time.Time
	End        time.Time
}

type renderContext struct {
	Days          []time.Time
	Rows          []renderRow
	DayCount      int
	TodayIndex    int
	HasActual     bool
	HasNotes      bool
	LiveReloadURL string
	CSS           template.CSS
}
