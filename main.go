package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"file-server/internal/app"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	// version is set at build time via -ldflags "-X main.version=..." for official releases.
	version = "dev"
	commit  = ""
	date    = ""
)

func getVersion() string {
	if version != "dev" && version != "" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}

		var rev string
		var modified bool
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					rev = setting.Value[:7]
				} else {
					rev = setting.Value
				}
			case "vcs.modified":
				if setting.Value == "true" {
					modified = true
				}
			}
		}

		if rev != "" {
			if modified {
				return fmt.Sprintf("dev-%s-dirty", rev)
			}
			return fmt.Sprintf("dev-%s", rev)
		}
	}

	return "dev"
}

func main() {
	var port int
	var showVersion bool

	flag.IntVar(&port, "port", 8000, "Preferred HTTP server port")
	flag.IntVar(&port, "p", 8000, "Preferred HTTP server port (shorthand)")
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")
	flag.BoolVar(&showVersion, "v", false, "Print version information and exit (shorthand)")
	flag.Parse()

	if showVersion {
		fmt.Printf("file-server version %s\n", getVersion())
		return
	}

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
