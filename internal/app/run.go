package app

import (
	"context"
	"fmt"
	"os"

	"ilaunch/internal/env"
	"ilaunch/internal/runner"
	"ilaunch/internal/system"

	tea "github.com/charmbracelet/bubbletea"
)

func RunInteractive(ctx context.Context) (int, error) {
	check, err := system.CheckEnvironment(ctx, system.ExecCommander{})
	if err != nil {
		return 1, fmt.Errorf("environment checks failed: %w", err)
	}
	model := NewModel(check)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithContext(ctx))
	final, err := program.Run()
	if err != nil {
		return 1, fmt.Errorf("run tui: %w", err)
	}
	m, ok := final.(Model)
	if !ok {
		return 1, fmt.Errorf("invalid final model type")
	}
	if m.err != nil {
		return m.exitCodeOrDefault(), m.err
	}
	return m.exitCodeOrDefault(), nil
}

func RunNonInteractive(ctx context.Context) (int, error) {
	check, err := system.CheckEnvironment(ctx, system.ExecCommander{})
	if err != nil {
		return 1, fmt.Errorf("environment checks failed: %w", err)
	}

	if _, err = os.Stat(".env"); os.IsNotExist(err) {
		if err = createEnvWithDefaults(); err != nil {
			return 1, fmt.Errorf("create env: %w", err)
		}
		fmt.Println("created .env from .env.example defaults")
	}

	r := runner.Runner{}
	if code, err := streamProcess(ctx, r, check.PackageMgr, "install"); err != nil {
		return code, err
	}
	if _, err = os.Stat(".git"); os.IsNotExist(err) {
		if code, err := streamProcess(ctx, r, "git", "init"); err != nil {
			return code, err
		}
		if code, err := streamProcess(ctx, r, "git", "add", "."); err != nil {
			return code, err
		}
		if code, err := streamProcess(ctx, r, "git", "commit", "-m", "Initial commit"); err != nil {
			return code, err
		}
	}
	return 0, nil
}

func streamProcess(ctx context.Context, r runner.Runner, name string, args ...string) (int, error) {
	fmt.Printf("$ %s %v\n", name, args)
	for ev := range r.Run(ctx, name, args...) {
		switch ev.Type {
		case runner.EventLine:
			fmt.Println(ev.Line)
		case runner.EventError:
			return 1, ev.Err
		case runner.EventDone:
			if ev.Err != nil || ev.ExitCode != 0 {
				return ev.ExitCode, fmt.Errorf("command failed: %w", ev.Err)
			}
		}
	}
	return 0, nil
}

func createEnvWithDefaults() error {
	file, err := os.Open(".env.example")
	if err != nil {
		return fmt.Errorf("open .env.example: %w", err)
	}
	defer file.Close()
	entries, err := env.ParseExample(file)
	if err != nil {
		return fmt.Errorf("parse .env.example: %w", err)
	}
	values := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.Default == "" {
			return fmt.Errorf("empty default value for key %s", e.Key)
		}
		values[e.Key] = e.Default
	}
	if err := env.WriteFile(".env", values); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}
	return nil
}

func (m Model) exitCodeOrDefault() int {
	if m.exitCode != 0 {
		return m.exitCode
	}
	return 0
}
