package app

import (
	"fmt"
	"os"
	"strings"

	"ilaunch/internal/env"

	tea "github.com/charmbracelet/bubbletea"
)

var menuItems = []string{
	"Create .env file",
	"Install dependencies",
	"Initialize git",
	"Run all",
	"Exit",
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(typed)
	case ProcessMsg:
		return m.handleProcessMsg(typed)
	case ErrorMsg:
		m.setError(typed.Err)
		return m, nil
	case ProgressMsg:
		if typed.Value > m.progress {
			m.progress = typed.Value
		}
		if m.progress >= 1 {
			m.running = false
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if k.String() == "ctrl+c" {
		m.cancel()
		m.exitCode = 130
		return m, tea.Quit
	}
	if k.String() == "esc" {
		if m.running {
			m.cancel()
			m.setError(fmt.Errorf("operation canceled"))
			return m, nil
		}
		if m.screen == ScreenEnvForm || m.screen == ScreenLogs {
			m.screen = ScreenMenu
			return m, nil
		}
		if m.screen == ScreenMenu || m.screen == ScreenError {
			return m, tea.Quit
		}
	}

	switch m.screen {
	case ScreenMenu:
		switch k.String() {
		case "up":
			if m.menuIndex > 0 {
				m.menuIndex--
			}
		case "down":
			if m.menuIndex < menuItemCount-1 {
				m.menuIndex++
			}
		case "enter":
			return m.handleMenuAction()
		}
	case ScreenEnvForm:
		return m.handleEnvFormInput(k)
	case ScreenLogs:
		switch k.String() {
		case "up":
			if m.scroll < len(m.logs)-1 {
				m.scroll++
			}
		case "down":
			if m.scroll > 0 {
				m.scroll--
			}
		}
	case ScreenError:
		if k.String() == "enter" {
			m.screen = ScreenMenu
		}
	}
	return m, nil
}

func (m Model) handleMenuAction() (tea.Model, tea.Cmd) {
	switch m.menuIndex {
	case 0:
		return m, m.beginCreateEnv()
	case 1:
		cmd := m.startProcess(m.checkResult.PackageMgr, "install")
		return m, cmd
	case 2:
		return m, m.startGitInit()
	case 3:
		return m, m.runAll()
	case 4:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m Model) startGitInit() tea.Cmd {
	if _, err := os.Stat(".git"); err == nil {
		m.addLog("git already initialized")
		return nil
	}
	m.enqueue([]string{"git", "init"}, []string{"git", "add", "."}, []string{"git", "commit", "-m", "Initial commit"})
	return m.startNextQueued()
}

func (m Model) runAll() tea.Cmd {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		if createErr := createEnvWithDefaults(); createErr != nil {
			m.setError(fmt.Errorf("create env defaults: %w", createErr))
			return nil
		}
		m.addLog(".env file created from defaults")
	}
	m.enqueue([]string{m.checkResult.PackageMgr, "install"})
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		m.enqueue([]string{"git", "init"}, []string{"git", "add", "."}, []string{"git", "commit", "-m", "Initial commit"})
	}
	return m.startNextQueued()
}

func (m Model) handleProcessMsg(msg ProcessMsg) (tea.Model, tea.Cmd) {
	ev := msg.Event
	switch ev.Type {
	case 0:
		m.addLog(ev.Line)
		if m.progress < 0.95 {
			m.progress += 0.02
		}
		return m, waitProcessEvent(m.processCh)
	case 1:
		m.running = false
		m.progress = 1
		if ev.Err != nil || ev.ExitCode != 0 {
			m.setError(fmt.Errorf("process failed (code %d): %w", ev.ExitCode, ev.Err))
			return m, nil
		}
		m.addLog("process completed successfully")
		if len(m.pending) > 0 {
			return m, m.startNextQueued()
		}
		return m, nil
	case 2:
		m.setError(ev.Err)
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleEnvFormInput(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.envEntries) == 0 {
		m.setError(fmt.Errorf(".env.example has no entries"))
		return m, nil
	}
	switch k.Type {
	case tea.KeyBackspace:
		if len(m.fieldInput) > 0 {
			m.fieldInput = m.fieldInput[:len(m.fieldInput)-1]
		}
	case tea.KeyRunes:
		m.fieldInput += string(k.Runes)
	case tea.KeyEnter:
		current := m.envEntries[m.fieldIndex]
		if strings.TrimSpace(m.fieldInput) == "" {
			m.setError(fmt.Errorf("value for %s cannot be empty", current.Key))
			return m, nil
		}
		m.envValues[current.Key] = strings.TrimSpace(m.fieldInput)
		m.fieldIndex++
		if m.fieldIndex >= len(m.envEntries) {
			if err := env.WriteFile(".env", m.envValues); err != nil {
				m.setError(fmt.Errorf("write .env: %w", err))
				return m, nil
			}
			m.addLog(".env file created")
			m.screen = ScreenMenu
			return m, nil
		}
		m.fieldInput = m.envEntries[m.fieldIndex].Default
	}
	return m, nil
}
