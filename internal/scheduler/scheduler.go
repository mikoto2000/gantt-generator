package scheduler

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"ganttgen/internal/calendar"
	"ganttgen/internal/model"
)

// Schedule resolves task dates respecting dependencies and workday rules.
func Schedule(tasks []model.Task) ([]model.Task, error) {
	if len(tasks) == 0 {
		return nil, errors.New("no tasks to schedule")
	}

	byName := make(map[string]model.Task, len(tasks))
	for i := range tasks {
		task := tasks[i]
		byName[task.Name] = task
	}

	indegree := make(map[string]int, len(tasks))
	graph := make(map[string][]string, len(tasks))
	for _, t := range tasks {
		indegree[t.Name] = len(t.DependsOn)
		for _, dep := range t.DependsOn {
			graph[dep] = append(graph[dep], t.Name)
		}
	}

	queue := make([]string, 0, len(tasks))
	for name, deg := range indegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue) // deterministic start order

	scheduled := make(map[string]model.Task, len(tasks))
	var ordered []model.Task

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		taskVal, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown task referenced in queue: %s", name)
		}
		scheduledTask, err := computeSchedule(taskVal, scheduled)
		if err != nil {
			return nil, err
		}
		scheduled[name] = scheduledTask
		ordered = append(ordered, scheduledTask)

		for _, successor := range graph[name] {
			indegree[successor]--
			if indegree[successor] == 0 {
				queue = append(queue, successor)
			}
		}
	}

	if len(ordered) != len(tasks) {
		return nil, errors.New("cyclic dependency detected")
	}
	return ordered, nil
}

func computeSchedule(task model.Task, scheduled map[string]model.Task) (model.Task, error) {
	var (
		start    modelTaskDate
		hasStart bool
	)

	if task.Start != nil {
		start = modelTaskDate{calendar.NextWorkday(*task.Start)}
		hasStart = true
	}

	if len(task.DependsOn) > 0 {
		var (
			latestEnd modelTaskDate
			seen      bool
		)
		for _, dep := range task.DependsOn {
			depTask, ok := scheduled[dep]
			if !ok {
				return model.Task{}, fmt.Errorf("dependency %q for task %q not scheduled", dep, task.Name)
			}
			if !seen || depTask.ComputedEnd.After(latestEnd.Time) {
				latestEnd = modelTaskDate{depTask.ComputedEnd}
				seen = true
			}
		}
		depStart := modelTaskDate{calendar.NextWorkdayAfter(latestEnd.Time)}
		if !hasStart || depStart.After(start.Time) {
			start = depStart
			hasStart = true
		}
	}

	if !hasStart {
		return model.Task{}, fmt.Errorf("task %q lacks a resolvable start date", task.Name)
	}

	var end modelTaskDate
	if task.End != nil {
		end = modelTaskDate{calendar.NextWorkday(*task.End)}
		if end.Before(start.Time) {
			return model.Task{}, fmt.Errorf("task %q ends before it can start", task.Name)
		}
	} else if task.DurationDays > 0 {
		end = modelTaskDate{calendar.AddWorkdays(start.Time, task.DurationDays-1)}
	} else {
		return model.Task{}, fmt.Errorf("task %q lacks duration or end", task.Name)
	}

	task.ComputedStart = start.Time
	task.ComputedEnd = end.Time
	return task, nil
}

type modelTaskDate struct {
	time.Time
}

func (d modelTaskDate) After(t time.Time) bool {
	return d.Time.After(t)
}

func (d modelTaskDate) Before(t time.Time) bool {
	return d.Time.Before(t)
}
