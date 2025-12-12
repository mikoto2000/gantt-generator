package renderer

import (
	"strings"
	"testing"
	"time"

	"ganttgen/internal/model"
)

func TestBuildHTMLRendersTasks(t *testing.T) {
	tasks := []model.Task{
		{
			Name:                "Task A",
			ComputedStart:       day(2024, time.June, 3),
			ComputedEnd:         day(2024, time.June, 5),
			ComputedActualStart: ptrTime(day(2024, time.June, 1)),
			ComputedActualEnd:   ptrTime(day(2024, time.June, 4)),
		},
		{
			Name:                "Task B",
			ComputedStart:       day(2024, time.June, 6),
			ComputedEnd:         day(2024, time.June, 6),
			ComputedActualStart: ptrTime(day(2024, time.June, 7)),
			ComputedActualEnd:   ptrTime(day(2024, time.June, 10)),
		},
	}

	html, err := BuildHTML(tasks, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Task A") || !strings.Contains(html, "Task B") {
		t.Fatalf("task names not rendered")
	}
	if !strings.Contains(html, "grid-column:3 / span 3") {
		t.Fatalf("expected span for Task A not found")
	}
	if !strings.Contains(html, "grid-column:6 / span 1") {
		t.Fatalf("expected span for Task B not found")
	}
	if !strings.Contains(html, "grid-column:1 / span 4") { // actual for Task A
		t.Fatalf("expected actual span for Task A not found")
	}
	if !strings.Contains(html, "legend-swatch actual") {
		t.Fatalf("actual legend not rendered")
	}
}

func TestBuildHTMLIncludesLiveReload(t *testing.T) {
	tasks := []model.Task{
		{Name: "A", ComputedStart: day(2024, time.June, 3), ComputedEnd: day(2024, time.June, 3)},
	}
	url := "http://localhost:35729/livereload"
	html, err := BuildHTML(tasks, url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "EventSource('"+url+"')") && !strings.Contains(html, "EventSource('http:\\/\\/localhost:35729\\/livereload')") {
		t.Fatalf("livereload script missing")
	}
}

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func ptrTime(t time.Time) *time.Time { return &t }
