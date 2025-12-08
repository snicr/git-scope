package tui

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust table height based on window size
		m.table.SetHeight(m.height - 6) // Leave room for header and footer

	case scanCompleteMsg:
		m.repos = msg.repos
		m.state = StateReady
		m.table.SetRows(reposToRows(m.repos))
		m.statusMsg = ""
		return m, nil

	case scanErrorMsg:
		m.state = StateError
		m.err = msg.err
		return m, nil

	case openEditorMsg:
		// Open the editor asynchronously
		return m, tea.ExecProcess(exec.Command(m.cfg.Editor, msg.path), nil)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if m.state == StateReady && len(m.repos) > 0 {
				// Get the selected repo
				selectedRow := m.table.Cursor()
				if selectedRow >= 0 && selectedRow < len(m.repos) {
					// Find the actual repo (repos are sorted in reposToRows)
					row := m.table.SelectedRow()
					if len(row) > 0 {
						// Find repo by name (first column)
						for _, repo := range m.repos {
							if repo.Name == row[0] {
								m.statusMsg = "Opening in " + m.cfg.Editor + "..."
								return m, func() tea.Msg {
									return openEditorMsg{path: repo.Path}
								}
							}
						}
					}
				}
			}

		case "r":
			// Rescan
			m.state = StateLoading
			m.statusMsg = "Rescanning..."
			return m, scanReposCmd(m.cfg)
		}
	}

	// Update the table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
