package renderer

import (
	"errors"
	"html/template"
	"sort"
	"time"

	"ganttgen/internal/calendar"
	"ganttgen/internal/model"
)

// BuildHTML prepares render data and returns the final HTML string.
func BuildHTML(tasks []model.Task) (string, error) {
	if len(tasks) == 0 {
		return "", errors.New("no tasks to render")
	}

	// Ensure deterministic order: start date then name.
	sorted := make([]model.Task, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].ComputedStart.Equal(sorted[j].ComputedStart) {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].ComputedStart.Before(sorted[j].ComputedStart)
	})

	minStart, maxEnd := sorted[0].ComputedStart, sorted[0].ComputedEnd
	for _, t := range sorted[1:] {
		if t.ComputedStart.Before(minStart) {
			minStart = t.ComputedStart
		}
		if t.ComputedEnd.After(maxEnd) {
			maxEnd = t.ComputedEnd
		}
	}
	for _, t := range sorted {
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

	var rendered []renderTask
	var hasActual bool
	for _, t := range sorted {
		startIdx := daysBetween(minStart, t.ComputedStart)
		span := daysBetween(t.ComputedStart, t.ComputedEnd) + 1
		rt := renderTask{
			Name:       t.Name,
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
		rendered = append(rendered, rt)
	}

	ctx := renderContext{
		Days:       days,
		Tasks:      rendered,
		DayCount:   len(days),
		TodayIndex: todayIndex,
		HasActual:  hasActual,
		CSS:        template.CSS(baseCSS()),
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
	StartIndex int
	Span       int
	Start      time.Time
	End        time.Time
	Actual     *renderActual
}

type renderActual struct {
	StartIndex int
	Span       int
	Start      time.Time
	End        time.Time
}

type renderContext struct {
	Days       []time.Time
	Tasks      []renderTask
	DayCount   int
	TodayIndex int
	HasActual  bool
	CSS        template.CSS
}
