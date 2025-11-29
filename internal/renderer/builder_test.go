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
			Name:          "Task A",
			ComputedStart: day(2024, time.June, 3),
			ComputedEnd:   day(2024, time.June, 5),
		},
		{
			Name:          "Task B",
			ComputedStart: day(2024, time.June, 6),
			ComputedEnd:   day(2024, time.June, 6),
		},
	}

	html, err := BuildHTML(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Task A") || !strings.Contains(html, "Task B") {
		t.Fatalf("task names not rendered")
	}
	if !strings.Contains(html, "grid-column:1 / span 3") {
		t.Fatalf("expected span for Task A not found")
	}
	if !strings.Contains(html, "grid-column:4 / span 1") {
		t.Fatalf("expected span for Task B not found")
	}
}

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}
