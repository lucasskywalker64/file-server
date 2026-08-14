package main

import (
	"flag"
	"fmt"
	"os"

	"file-server/internal/app"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8000, "Preferred HTTP server port")
	flag.IntVar(&port, "p", 8000, "Preferred HTTP server port (shorthand)")
	flag.Parse()

	initialDir := "."
	if flag.NArg() > 0 {
		initialDir = flag.Arg(0)
	}

	model, err := app.NewModel(initialDir, port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI program: %v\n", err)
		os.Exit(1)
	}
}
