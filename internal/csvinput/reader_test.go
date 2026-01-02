package csvinput

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func TestReadValidCSV(t *testing.T) {
	content := `タスク名,開始,終了,期間,依存,実績開始,実績終了,実績期間
Planning,2024-06-03,,5d,,2024-06-03,,5d
Design,,,4d,Planning,2024-06-11,2024-06-17,
`
	dir := t.TempDir()
	path := filepath.Join(dir, "input.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, _, _, err := Read(path)
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
	if tasks[0].ComputedActualEnd == nil || !tasks[0].ComputedActualEnd.Equal(time.Date(2024, 6, 7, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("unexpected actual end for Planning: %v", tasks[0].ComputedActualEnd)
	}
	if tasks[1].ComputedActualStart == nil || !tasks[1].ComputedActualStart.Equal(time.Date(2024, 6, 11, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("unexpected actual start for Design: %v", tasks[1].ComputedActualStart)
	}
}

func TestReadStatusCancelled(t *testing.T) {
	content := `name,start,end,duration,depends_on,status
CancelledTask,2024-06-03,,1d,,cancelled
`
	dir := t.TempDir()
	path := filepath.Join(dir, "status.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, _, _, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Status != "cancelled" {
		t.Fatalf("unexpected status: %q", tasks[0].Status)
	}
	if !tasks[0].IsCancelled() {
		t.Fatalf("expected task to be cancelled")
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

	_, _, _, err := Read(path)
	if err == nil || !strings.Contains(err.Error(), "duplicate task name") {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
}

func TestReadDetectsShiftJIS(t *testing.T) {
	content := "タスク名,開始,終了,期間,依存\n計画,2024-06-03,,2d,\n"
	encoded, _, err := transform.String(japanese.ShiftJIS.NewEncoder(), content)
	if err != nil {
		t.Fatalf("encode to Shift_JIS: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "sjis.csv")
	if err := os.WriteFile(path, []byte(encoded), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, _, _, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "計画" {
		t.Fatalf("decoded task name mismatch: %s", tasks[0].Name)
	}
}

func TestReadAcceptsSlashSeparatedDates(t *testing.T) {
	content := `name,start,end,duration,depends_on
SlashDate,2024/06/03,,3d,
`
	dir := t.TempDir()
	path := filepath.Join(dir, "slash.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, _, _, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Start == nil || tasks[0].Start.Format("2006-01-02") != "2024-06-03" {
		t.Fatalf("slash date not parsed correctly: %+v", tasks[0].Start)
	}
}

func TestReadAcceptsNonPaddedDates(t *testing.T) {
	content := `name,start,end,duration,depends_on
NonPadded,2024-6-3,,2d,
SlashNonPadded,2024/6/3,,1d,
`
	dir := t.TempDir()
	path := filepath.Join(dir, "nonpadded.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, _, _, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	for _, tc := range tasks {
		if tc.Start == nil || tc.Start.Format("2006-01-02") != "2024-06-03" {
			t.Fatalf("non-padded date not normalized: %+v", tc.Start)
		}
	}
}

func TestReadIgnoresEmptyRows(t *testing.T) {
	content := `name,start,end,duration,depends_on
Filled,2024-06-03,,1d,
,,,, 
`
	dir := t.TempDir()
	path := filepath.Join(dir, "emptyrow.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, _, _, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "Filled" {
		t.Fatalf("unexpected task name: %s", tasks[0].Name)
	}
}

func TestReadProgressColumn(t *testing.T) {
	content := `name,start,end,duration,depends_on,progress
Task,2024-06-03,,1d,,45%
`
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, _, hasProgress, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasProgress {
		t.Fatalf("expected progress column to be detected")
	}
	if len(tasks) != 1 || tasks[0].ProgressPercent == nil || *tasks[0].ProgressPercent != 45 {
		t.Fatalf("unexpected progress percent: %#v", tasks)
	}
}

func TestReadCapturesCustomColumns(t *testing.T) {
	content := `name,start,end,duration,depends_on,担当,priority
Task,2024-06-03,,1d,,Alice,High
`
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	tasks, customCols, _, err := Read(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(customCols) != 2 || customCols[0] != "担当" || customCols[1] != "priority" {
		t.Fatalf("unexpected custom columns: %#v", customCols)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if len(tasks[0].CustomValues) != 2 || tasks[0].CustomValues[0] != "Alice" || tasks[0].CustomValues[1] != "High" {
		t.Fatalf("unexpected custom values: %#v", tasks[0].CustomValues)
	}
}
