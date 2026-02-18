package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	focusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func (m Model) View() string {
	switch m.screen {
	case ScreenMenu:
		return m.viewMenu()
	case ScreenEnvForm:
		return m.viewEnvForm()
	case ScreenLogs:
		return m.viewLogs()
	case ScreenError:
		return m.viewError()
	default:
		return ""
	}
}

func (m Model) viewMenu() string {
	rows := []string{titleStyle.Render("iLaunch — Project Bootstrap TUI"), mutedStyle.Render(fmt.Sprintf("Node %s | %s", m.checkResult.NodeVersion, m.checkResult.PackageMgr)), ""}
	for i, item := range menuItems {
		prefix := "  "
		style := lipgloss.NewStyle()
		if m.menuIndex == i {
			prefix = "➜ "
			style = focusStyle
		}
		rows = append(rows, style.Render(prefix+item))
	}
	rows = append(rows, "", mutedStyle.Render("↑/↓ navigate • Enter select • Esc exit"))
	return boxStyle.Width(m.width - 4).Render(strings.Join(rows, "\n"))
}

func (m Model) viewEnvForm() string {
	entry := m.envEntries[m.fieldIndex]
	rows := []string{
		titleStyle.Render("Create .env file"),
		mutedStyle.Render(fmt.Sprintf("Field %d/%d", m.fieldIndex+1, len(m.envEntries))),
		"",
		focusStyle.Render(entry.Key + "=" + m.fieldInput),
		"",
		mutedStyle.Render("Type value and press Enter • Esc back"),
	}
	return boxStyle.Width(m.width - 4).Render(strings.Join(rows, "\n"))
}

func (m Model) viewLogs() string {
	rows := []string{titleStyle.Render("Process logs"), progressBar(m.progress, m.width-12), ""}
	maxRows := m.height - 8
	if maxRows < 5 {
		maxRows = 5
	}
	start := len(m.logs) - maxRows - m.scroll
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > len(m.logs) {
		end = len(m.logs)
	}
	for _, line := range m.logs[start:end] {
		rows = append(rows, line)
	}
	rows = append(rows, "", mutedStyle.Render("↑/↓ scroll • Esc back"))
	return boxStyle.Width(m.width - 4).Render(strings.Join(rows, "\n"))
}

func (m Model) viewError() string {
	message := "unknown error"
	if m.err != nil {
		message = m.err.Error()
	}
	rows := []string{
		errStyle.Render("Error"),
		message,
		"",
		mutedStyle.Render("Enter: back to menu • Esc: quit"),
	}
	return boxStyle.Width(m.width - 4).Render(strings.Join(rows, "\n"))
}

func progressBar(progress float64, width int) string {
	if width < 10 {
		width = 10
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	barWidth := width - 8
	filled := int(progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	return fmt.Sprintf("[%s%s] %3.0f%%", strings.Repeat("█", filled), strings.Repeat("░", barWidth-filled), progress*100)
}
