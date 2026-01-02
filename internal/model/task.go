package model

import (
	"strings"
	"time"
)

// Task represents a single CSV-defined task and its computed schedule.
type Task struct {
	Name                string
	IsHeading           bool
	DisplayOnly         bool
	Notes               string
	Status              string
	CustomValues        []string
	Start               *time.Time
	End                 *time.Time
	DurationDays        int
	ActualStart         *time.Time
	ActualEnd           *time.Time
	ActualDurationDays  int
	DependsOn           []string
	ComputedStart       time.Time
	ComputedEnd         time.Time
	ComputedActualStart *time.Time
	ComputedActualEnd   *time.Time
}

// HasStart returns true when an absolute start date was provided.
func (t Task) HasStart() bool {
	return t.Start != nil
}

// HasEnd returns true when an absolute end date was provided.
func (t Task) HasEnd() bool {
	return t.End != nil
}

// HasDuration returns true when a duration was provided.
func (t Task) HasDuration() bool {
	return t.DurationDays > 0
}

// HasActual returns true when any actual-related date exists.
func (t Task) HasActual() bool {
	return t.ComputedActualStart != nil && t.ComputedActualEnd != nil
}

// IsCancelled reports whether the task is marked as cancelled by status.
func (t Task) IsCancelled() bool {
	status := strings.TrimSpace(strings.ToLower(t.Status))
	return status == "cancelled" || status == "中止"
}

// IsCompleted reports whether the task is marked as completed by status.
func (t Task) IsCompleted() bool {
	status := strings.TrimSpace(strings.ToLower(t.Status))
	return status == "completed" || status == "完了"
}
