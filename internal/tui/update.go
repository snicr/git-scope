package tui

import (
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust table height based on window size
		tableHeight := m.height - 12
		if tableHeight < 5 {
			tableHeight = 5
		}
		m.table.SetHeight(tableHeight)

	case scanCompleteMsg:
		m.repos = msg.repos
		m.state = StateReady
		m.updateTable()
		m.statusMsg = ""
		return m, nil

	case scanErrorMsg:
		m.state = StateError
		m.err = msg.err
		return m, nil

	case openEditorMsg:
		c := exec.Command(m.cfg.Editor, msg.path)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return editorClosedMsg{err: err}
			}
			return editorClosedMsg{}
		})

	case editorClosedMsg:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
		} else {
			m.statusMsg = ""
		}
		return m, scanReposCmd(m.cfg)

	case tea.KeyMsg:
		// Handle search mode separately
		if m.state == StateSearching {
			return m.handleSearchMode(msg)
		}
		
		// Normal mode key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "/":
			// Enter search mode
			if m.state == StateReady {
				m.state = StateSearching
				m.textInput.Focus()
				m.textInput.SetValue(m.searchQuery)
				return m, textinput.Blink
			}

		case "enter":
			if m.state == StateReady {
				repo := m.GetSelectedRepo()
				if repo != nil {
					m.statusMsg = "Opening " + repo.Name + " in " + m.cfg.Editor + "..."
					return m, func() tea.Msg {
						return openEditorMsg{path: repo.Path}
					}
				}
			}

		case "r":
			m.state = StateLoading
			m.statusMsg = "Rescanning..."
			return m, scanReposCmd(m.cfg)

		case "f":
			// Cycle through filter modes
			if m.state == StateReady {
				m.filterMode = (m.filterMode + 1) % 3
				m.updateTable()
				m.statusMsg = "Filter: " + m.GetFilterModeName()
				return m, nil
			}

		case "s":
			if m.state == StateReady {
				m.sortMode = (m.sortMode + 1) % 4
				m.updateTable()
				m.statusMsg = "Sorted by: " + m.GetSortModeName()
				return m, nil
			}

		case "1":
			if m.state == StateReady {
				m.sortMode = SortByDirty
				m.updateTable()
				m.statusMsg = "Sorted by: Dirty First"
				return m, nil
			}

		case "2":
			if m.state == StateReady {
				m.sortMode = SortByName
				m.updateTable()
				m.statusMsg = "Sorted by: Name"
				return m, nil
			}

		case "3":
			if m.state == StateReady {
				m.sortMode = SortByBranch
				m.updateTable()
				m.statusMsg = "Sorted by: Branch"
				return m, nil
			}

		case "4":
			if m.state == StateReady {
				m.sortMode = SortByLastCommit
				m.updateTable()
				m.statusMsg = "Sorted by: Recent"
				return m, nil
			}

		case "c":
			// Clear search and filters
			if m.state == StateReady {
				m.searchQuery = ""
				m.filterMode = FilterAll
				m.updateTable()
				m.statusMsg = "Filters cleared"
				return m, nil
			}

		case "e":
			if m.state == StateReady {
				m.statusMsg = "Editor: " + m.cfg.Editor + " (change in ~/.config/git-scope/config.yml)"
				return m, nil
			}
		}
	}

	// Update the table
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// handleSearchMode handles key events when in search mode
func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel search, keep previous query
		m.state = StateReady
		m.textInput.Blur()
		return m, nil
		
	case "enter":
		// Apply search
		m.searchQuery = m.textInput.Value()
		m.state = StateReady
		m.textInput.Blur()
		m.updateTable()
		if m.searchQuery != "" {
			m.statusMsg = "Searching: " + m.searchQuery
		} else {
			m.statusMsg = "Search cleared"
		}
		return m, nil
		
	case "ctrl+c":
		return m, tea.Quit
	}
	
	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	
	// Live search as you type
	m.searchQuery = m.textInput.Value()
	m.updateTable()
	
	return m, cmd
}

// editorClosedMsg is sent when the editor process closes
type editorClosedMsg struct {
	err error
}
