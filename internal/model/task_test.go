package model

import "testing"

func TestIsCompleted(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{status: "completed", want: true},
		{status: "完了", want: true},
		{status: "in progress", want: false},
	}
	for _, tc := range cases {
		task := Task{Status: tc.status}
		if got := task.IsCompleted(); got != tc.want {
			t.Fatalf("status %q: expected %v, got %v", tc.status, tc.want, got)
		}
	}
}
