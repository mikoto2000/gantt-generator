package csvinput

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadValidCSV(t *testing.T) {
	content := `name,start,end,duration,depends_on
Planning,2024-06-03,,5d,
Design,,,4d,Planning
`
	dir := t.TempDir()
	path := filepath.Join(dir, "input.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Name != "Planning" || tasks[1].Name != "Design" {
		t.Fatalf("unexpected task order or names: %#v", tasks)
	}
	if tasks[0].Start == nil || !tasks[0].Start.Equal(time.Date(2024, 6, 3, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("unexpected start date for Planning")
	}
	if tasks[1].DurationDays != 4 {
		t.Fatalf("unexpected duration for Design: %d", tasks[1].DurationDays)
	}
	if got := tasks[1].DependsOn; len(got) != 1 || got[0] != "Planning" {
		t.Fatalf("unexpected depends_on: %#v", got)
	}
}

func TestReadDuplicateName(t *testing.T) {
	content := `name,start,end,duration,depends_on
A,2024-06-03,,1d,
A,2024-06-04,,1d,
`
	dir := t.TempDir()
	path := filepath.Join(dir, "dup.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := Read(path)
	if err == nil || !strings.Contains(err.Error(), "duplicate task name") {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
}
