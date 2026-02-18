package runner

import (
	"context"
	"runtime"
	"testing"
)

func TestRunnerRun(t *testing.T) {
	r := Runner{}
	ctx := context.Background()
	name := "sh"
	args := []string{"-c", "echo hello"}
	if runtime.GOOS == "windows" {
		name = "cmd"
		args = []string{"/C", "echo hello"}
	}
	events := r.Run(ctx, name, args...)
	seenLine := false
	seenDone := false
	for ev := range events {
		switch ev.Type {
		case EventLine:
			if ev.Line != "" {
				seenLine = true
			}
		case EventDone:
			if ev.ExitCode != 0 {
				t.Fatalf("expected exit code 0, got %d", ev.ExitCode)
			}
			seenDone = true
		}
	}
	if !seenLine || !seenDone {
		t.Fatalf("expected line and done events, got line=%v done=%v", seenLine, seenDone)
	}
}
