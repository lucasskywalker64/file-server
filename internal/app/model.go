package app

import (
	"fmt"
	"os"

	"file-server/internal/fs"
	"file-server/internal/network"
	"file-server/internal/server"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	StateNavigating State = iota
	StateServing
)

// LogEventMsg carries an HTTP request log event to the Bubble Tea update loop.
type LogEventMsg server.LogEvent

// ErrMsg carries error messages to the TUI.
type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

// Model represents the overall state of the file-server TUI application.
type Model struct {
	State          State
	Cwd            string
	Entries        []fs.Entry
	Cursor         int
	ViewportStart  int
	ViewportHeight int
	Width          int
	Height         int

	PreferredPort int

	Server     *server.Server
	ServedFile fs.Entry
	LogChan    chan server.LogEvent
	Logs       []server.LogEvent
	LogOffset  int // 0 means auto-scroll bottom (newest logs), >0 scrolls up into history
	LANIPs     []string

	Err error
}

// NewModel initializes the application model given an initial directory and preferred port.
func NewModel(initialDir string, preferredPort int) (Model, error) {
	if initialDir == "" {
		var err error
		initialDir, err = os.Getwd()
		if err != nil {
			return Model{}, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	entries, absCwd, err := fs.ReadDirectory(initialDir)
	if err != nil {
		return Model{}, err
	}

	lanIPs := network.GetLANIPs()

	return Model{
		State:          StateNavigating,
		Cwd:            absCwd,
		Entries:        entries,
		Cursor:         0,
		ViewportStart:  0,
		ViewportHeight: 15, // Default fallback before window size msg
		Width:          80,
		Height:         24,
		PreferredPort:  preferredPort,
		LANIPs:         lanIPs,
	}, nil
}

// Init initializes Bubble Tea commands.
func (m Model) Init() tea.Cmd {
	return nil
}

// listenForLog returns a tea.Cmd that waits for a log event on logChan.
func listenForLog(logChan chan server.LogEvent) tea.Cmd {
	return func() tea.Msg {
		if logChan == nil {
			return nil
		}
		evt, ok := <-logChan
		if !ok {
			return nil
		}
		return LogEventMsg(evt)
	}
}
