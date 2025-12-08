package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Padding(0, 1).
			MarginBottom(1)

	// Subtitle for repo count
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	// Status message style
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			MarginTop(1)

	// Loading style
	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
)

// View renders the TUI
func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("git-scope"))
	b.WriteString("  ")

	switch m.state {
	case StateLoading:
		b.WriteString(loadingStyle.Render("Scanning repositories..."))
		b.WriteString("\n\n")
		b.WriteString("Searching for git repos in:\n")
		for _, root := range m.cfg.Roots {
			b.WriteString("  • " + root + "\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Press q to quit"))

	case StateError:
		b.WriteString(errorStyle.Render("Error"))
		b.WriteString("\n\n")
		if m.err != nil {
			b.WriteString(errorStyle.Render(m.err.Error()))
		}
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press q to quit • r to retry"))

	case StateReady:
		// Subtitle with repo count
		dirtyCount := 0
		for _, r := range m.repos {
			if r.Status.IsDirty {
				dirtyCount++
			}
		}
		subtitle := fmt.Sprintf("%d repos found", len(m.repos))
		if dirtyCount > 0 {
			subtitle += fmt.Sprintf(" (%d dirty)", dirtyCount)
		}
		b.WriteString(subtitleStyle.Render(subtitle))
		b.WriteString("\n\n")

		// Table
		b.WriteString(m.table.View())
		b.WriteString("\n")

		// Status message
		if m.statusMsg != "" {
			b.WriteString(statusStyle.Render(m.statusMsg))
			b.WriteString("\n")
		}

		// Help
		b.WriteString(helpStyle.Render("↑/↓ navigate • enter open in editor • r rescan • q quit"))
	}

	return b.String()
}
