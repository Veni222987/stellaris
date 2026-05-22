package protocol

import "testing"

func TestTaskTopic(t *testing.T) {
	got := TaskTopic("ab12cd34ef567890", "planet-uuid-1")
	want := "stellaris/ab12cd34ef567890/task/planet-uuid-1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResultTopic(t *testing.T) {
	got := ResultTopic("ab12cd34ef567890", "task-uuid-1")
	want := "stellaris/ab12cd34ef567890/result/task-uuid-1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
