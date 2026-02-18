package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type EventType int

const (
	EventLine EventType = iota
	EventDone
	EventError
)

type Event struct {
	Type     EventType
	Line     string
	ExitCode int
	Err      error
}

type Runner struct{}

func (Runner) Run(ctx context.Context, name string, args ...string) <-chan Event {
	ch := make(chan Event)
	go func() {
		defer close(ch)
		cmd := exec.CommandContext(ctx, name, args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			ch <- Event{Type: EventError, Err: fmt.Errorf("open stdout pipe: %w", err)}
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			ch <- Event{Type: EventError, Err: fmt.Errorf("open stderr pipe: %w", err)}
			return
		}
		if err = cmd.Start(); err != nil {
			ch <- Event{Type: EventError, Err: fmt.Errorf("start process %s: %w", name, err)}
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)
		readPipe := func(r io.Reader) {
			defer wg.Done()
			s := bufio.NewScanner(r)
			for s.Scan() {
				ch <- Event{Type: EventLine, Line: s.Text()}
			}
			if scanErr := s.Err(); scanErr != nil {
				ch <- Event{Type: EventError, Err: fmt.Errorf("read process output: %w", scanErr)}
			}
		}
		go readPipe(stdout)
		go readPipe(stderr)
		wg.Wait()

		err = cmd.Wait()
		exitCode := cmd.ProcessState.ExitCode()
		if err != nil {
			ch <- Event{Type: EventDone, ExitCode: exitCode, Err: fmt.Errorf("wait process %s: %w", name, err)}
			return
		}
		ch <- Event{Type: EventDone, ExitCode: exitCode}
	}()
	return ch
}
