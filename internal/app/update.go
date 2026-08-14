package app

import (
	"file-server/internal/fs"
	"file-server/internal/server"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles state transitions and key/window events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.ViewportHeight = msg.Height - 7
		if m.ViewportHeight < 5 {
			m.ViewportHeight = 5
		}
		m.adjustViewport()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.Server != nil {
				_ = m.Server.Stop()
				m.Server = nil
				m.LogChan = nil
			}
			return m, tea.Quit

		case "esc":
			if m.State == StateServing {
				if m.Server != nil {
					_ = m.Server.Stop()
					m.Server = nil
					m.LogChan = nil
				}
				m.State = StateNavigating
				m.LogOffset = 0
				m.Err = nil
				return m, nil
			}
		}

		switch m.State {
		case StateNavigating:
			return m.updateNavigating(msg)
		case StateServing:
			return m.updateServing(msg)
		}

	case LogEventMsg:
		m.Logs = append(m.Logs, server.LogEvent(msg))
		if len(m.Logs) > 100 {
			m.Logs = m.Logs[len(m.Logs)-100:]
		}
		if m.State == StateServing && m.LogChan != nil {
			return m, listenForLog(m.LogChan)
		}

	case ErrMsg:
		m.Err = msg.Err
	}

	return m, nil
}

func (m Model) updateNavigating(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.Entries) - 1

	switch msg.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			m.adjustViewport()
		}

	case "down", "j":
		if m.Cursor < maxIdx {
			m.Cursor++
			m.adjustViewport()
		}

	case "left", "h":
		// Navigate to parent directory if available
		if len(m.Entries) > 0 && m.Entries[0].IsParent {
			entries, absPath, err := fs.ReadDirectory(m.Entries[0].Path)
			if err != nil {
				m.Err = err
				return m, nil
			}
			m.Cwd = absPath
			m.Entries = entries
			m.Cursor = 0
			m.ViewportStart = 0
			m.Err = nil
		}

	case "r":
		// Refresh current directory
		entries, absPath, err := fs.ReadDirectory(m.Cwd)
		if err != nil {
			m.Err = err
			return m, nil
		}
		m.Cwd = absPath
		m.Entries = entries
		m.adjustViewport()
		m.Err = nil

	case "pgup", "b":
		m.Cursor -= m.ViewportHeight
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		m.adjustViewport()

	case "pgdown", "f":
		m.Cursor += m.ViewportHeight
		if m.Cursor > maxIdx {
			m.Cursor = maxIdx
		}
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		m.adjustViewport()

	case "home", "g":
		m.Cursor = 0
		m.adjustViewport()

	case "end", "G":
		if maxIdx >= 0 {
			m.Cursor = maxIdx
			m.adjustViewport()
		}

	case "enter", "right", "l":
		if len(m.Entries) == 0 || m.Cursor < 0 || m.Cursor >= len(m.Entries) {
			return m, nil
		}

		entry := m.Entries[m.Cursor]
		if entry.IsDir {
			entries, absPath, err := fs.ReadDirectory(entry.Path)
			if err != nil {
				m.Err = err
				return m, nil
			}
			m.Cwd = absPath
			m.Entries = entries
			m.Cursor = 0
			m.ViewportStart = 0
			m.Err = nil
		} else {
			// Stop previous server if active
			if m.Server != nil {
				_ = m.Server.Stop()
				m.Server = nil
				m.LogChan = nil
			}

			m.LogChan = make(chan server.LogEvent, 100)
			srv, err := server.StartServer(entry.Path, m.PreferredPort, m.LogChan)
			if err != nil {
				m.Err = err
				return m, nil
			}

			m.Server = srv
			m.ServedFile = entry
			m.State = StateServing
			m.Logs = nil
			m.LogOffset = 0
			m.Err = nil

			return m, listenForLog(m.LogChan)
		}
	}

	return m, nil
}

func (m Model) updateServing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxOffset := len(m.Logs) - 1
	if maxOffset < 0 {
		maxOffset = 0
	}

	switch msg.String() {
	case "up", "k":
		if m.LogOffset < maxOffset {
			m.LogOffset++
		}
	case "down", "j":
		if m.LogOffset > 0 {
			m.LogOffset--
		}
	case "home", "g":
		m.LogOffset = maxOffset
	case "end", "G":
		m.LogOffset = 0
	}
	return m, nil
}

func (m *Model) adjustViewport() {
	if len(m.Entries) == 0 {
		m.Cursor = 0
		m.ViewportStart = 0
		return
	}

	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(m.Entries) {
		m.Cursor = len(m.Entries) - 1
	}

	if m.Cursor < m.ViewportStart {
		m.ViewportStart = m.Cursor
	}

	if m.Cursor >= m.ViewportStart+m.ViewportHeight {
		m.ViewportStart = m.Cursor - m.ViewportHeight + 1
	}

	if m.ViewportStart < 0 {
		m.ViewportStart = 0
	}
}
