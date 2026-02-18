package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"ilaunch/internal/env"
	"ilaunch/internal/runner"
	"ilaunch/internal/system"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxLogLines      = 300
	menuItemCount    = 5
	defaultWinWidth  = 100
	defaultWinHeight = 30
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenEnvForm
	ScreenLogs
	ScreenError
)

type ProcessMsg struct{ Event runner.Event }
type ErrorMsg struct{ Err error }
type ProgressMsg struct{ Value float64 }

type Model struct {
	screen      Screen
	menuIndex   int
	envEntries  []env.Entry
	envValues   map[string]string
	fieldIndex  int
	fieldInput  string
	logs        []string
	scroll      int
	progress    float64
	running     bool
	err         error
	width       int
	height      int
	checkResult system.CheckResult
	runner      runner.Runner
	processCh   <-chan runner.Event
	ctx         context.Context
	cancel      context.CancelFunc
	exitCode    int
	pending     [][]string
}

func NewModel(check system.CheckResult) Model {
	ctx, cancel := context.WithCancel(context.Background())
	return Model{
		screen:      ScreenMenu,
		envValues:   map[string]string{},
		width:       defaultWinWidth,
		height:      defaultWinHeight,
		checkResult: check,
		runner:      runner.Runner{},
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) addLog(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > maxLogLines {
		m.logs = m.logs[len(m.logs)-maxLogLines:]
	}
}

func (m *Model) setError(err error) {
	m.err = err
	m.screen = ScreenError
	m.exitCode = 1
}

func (m *Model) startProcess(name string, args ...string) tea.Cmd {
	m.running = true
	m.progress = 0.1
	m.screen = ScreenLogs
	m.addLog(fmt.Sprintf("$ %s %v", name, args))
	m.processCh = m.runner.Run(m.ctx, name, args...)
	return waitProcessEvent(m.processCh)
}

func waitProcessEvent(ch <-chan runner.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return ProgressMsg{Value: 1.0}
		}
		return ProcessMsg{Event: ev}
	}
}

func (m *Model) beginCreateEnv() tea.Cmd {
	examplePath := filepath.Join(".env.example")
	file, err := os.Open(examplePath)
	if err != nil {
		m.setError(fmt.Errorf("open .env.example: %w", err))
		return nil
	}
	defer file.Close()
	entries, err := env.ParseExample(file)
	if err != nil {
		m.setError(fmt.Errorf("parse .env.example: %w", err))
		return nil
	}
	m.envEntries = entries
	m.envValues = make(map[string]string, len(entries))
	m.fieldIndex = 0
	m.fieldInput = ""
	if len(entries) > 0 {
		m.fieldInput = entries[0].Default
	}
	m.screen = ScreenEnvForm
	return nil
}

func (m *Model) enqueue(commands ...[]string) {
	m.pending = append(m.pending, commands...)
}

func (m *Model) startNextQueued() tea.Cmd {
	if len(m.pending) == 0 {
		return nil
	}
	next := m.pending[0]
	m.pending = m.pending[1:]
	if len(next) == 0 {
		return m.startNextQueued()
	}
	name := next[0]
	args := []string{}
	if len(next) > 1 {
		args = next[1:]
	}
	return m.startProcess(name, args...)
}
