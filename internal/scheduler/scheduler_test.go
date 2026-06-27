package scheduler

import (
	"testing"
	"time"

	"ganttgen/internal/model"
)

func d(y int, m time.Month, day int) time.Time {
	return time.Date(y, m, day, 0, 0, 0, 0, time.Local)
}

func TestScheduleWithDependenciesAndWeekends(t *testing.T) {
	tasks := []model.Task{
		{
			Name:         "Planning",
			Start:        ptrTime(d(2024, time.June, 7)), // Friday
			DurationDays: 2,                              // spans Fri + Mon
		},
		{
			Name:         "Build",
			DependsOn:    []string{"Planning"},
			DurationDays: 3,
		},
	}

	got, err := Schedule(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}

	planning := findTask(t, got, "Planning")
	build := findTask(t, got, "Build")

	if !planning.ComputedStart.Equal(d(2024, time.June, 7)) {
		t.Fatalf("planning start mismatch: %v", planning.ComputedStart)
	}
	if !planning.ComputedEnd.Equal(d(2024, time.June, 10)) { // skips weekend
		t.Fatalf("planning end mismatch: %v", planning.ComputedEnd)
	}

	if !build.ComputedStart.Equal(d(2024, time.June, 11)) { // next workday after planning end
		t.Fatalf("build start mismatch: %v", build.ComputedStart)
	}
	if !build.ComputedEnd.Equal(d(2024, time.June, 13)) {
		t.Fatalf("build end mismatch: %v", build.ComputedEnd)
	}
}

func TestScheduleDetectsCycle(t *testing.T) {
	tasks := []model.Task{
		{Name: "A", DurationDays: 1, DependsOn: []string{"B"}},
		{Name: "B", DurationDays: 1, DependsOn: []string{"A"}},
	}
	if _, err := Schedule(tasks); err == nil {
		t.Fatalf("expected cycle detection error")
	}
}

func TestScheduleAllowsUnknownDependencyAsUnscheduled(t *testing.T) {
	tasks := []model.Task{
		{Name: "Blocked", DurationDays: 2, DependsOn: []string{"Missing"}},
	}

	got, err := Schedule(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	blocked := findTask(t, got, "Blocked")
	if !blocked.Unscheduled {
		t.Fatalf("expected task with unknown dependency to be unscheduled")
	}
}

func TestScheduleIgnoresUnknownDependencyWhenStartExists(t *testing.T) {
	tasks := []model.Task{
		{Name: "Started", Start: ptrTime(d(2024, time.June, 3)), DurationDays: 2, DependsOn: []string{"Missing"}},
	}

	got, err := Schedule(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	started := findTask(t, got, "Started")
	if started.Unscheduled {
		t.Fatalf("task with explicit start should be scheduled")
	}
	if !started.ComputedStart.Equal(d(2024, time.June, 3)) {
		t.Fatalf("start mismatch: %v", started.ComputedStart)
	}
	if !started.ComputedEnd.Equal(d(2024, time.June, 4)) {
		t.Fatalf("end mismatch: %v", started.ComputedEnd)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func findTask(t *testing.T, tasks []model.Task, name string) model.Task {
	t.Helper()
	for _, task := range tasks {
		if task.Name == name {
			return task
		}
	}
	t.Fatalf("task %s not found", name)
	return model.Task{}
}
