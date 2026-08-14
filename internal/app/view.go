package app

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.Color("#7D56F4") // Vibrant purple
	secondaryColor = lipgloss.Color("#04B575") // Vibrant green
	accentColor    = lipgloss.Color("#00D7FF") // Bright cyan
	textColor      = lipgloss.Color("#FAFAFA") // Off-white
	mutedColor     = lipgloss.Color("#626262") // Muted gray
	highlightBg    = lipgloss.Color("#352370") // Purple background highlight
	errorColor     = lipgloss.Color("#FF5F5F") // Soft red
	warningColor   = lipgloss.Color("#FFD700") // Gold

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 1)

	statusLiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(secondaryColor).
			Padding(0, 1)

	pathStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(highlightBg)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(textColor)

	dirIconStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	fileIconStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	parentIconStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	cardBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	urlBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	errorBoxStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	secondaryTextStyle = lipgloss.NewStyle().Foreground(secondaryColor)
	accentTextStyle    = lipgloss.NewStyle().Foreground(accentColor)
	warningTextStyle   = lipgloss.NewStyle().Foreground(warningColor)
)

// View renders the TUI screen based on the current state.
func (m Model) View() string {
	if m.State == StateServing {
		return m.viewServing()
	}
	return m.viewNavigating()
}

func (m Model) viewNavigating() string {
	var s string

	// Header
	header := titleStyle.Render(" TUI FILE SERVER ") + " " + pathStyle.Render(m.Cwd)
	s += header + "\n\n"

	if m.Err != nil {
		s += errorBoxStyle.Render("Error: "+m.Err.Error()) + "\n"
	}

	// File List Viewport
	endIdx := m.ViewportStart + m.ViewportHeight
	if endIdx > len(m.Entries) {
		endIdx = len(m.Entries)
	}

	if len(m.Entries) == 0 {
		s += mutedStyle.Render("  (empty directory)") + "\n"
	} else {
		for i := m.ViewportStart; i < endIdx; i++ {
			entry := m.Entries[i]

			var icon string
			if entry.IsParent {
				icon = parentIconStyle.Render("⬆️  ")
			} else if entry.IsDir {
				icon = dirIconStyle.Render("📁 ")
			} else {
				icon = fileIconStyle.Render("📄 ")
			}

			nameStr := entry.Name
			if entry.IsDir && !entry.IsParent {
				nameStr += "/"
			}

			sizeStr := ""
			if !entry.IsDir {
				sizeStr = mutedStyle.Render(fmt.Sprintf(" (%s)", formatSize(entry.Size)))
			}

			line := fmt.Sprintf("%s %s%s", icon, nameStr, sizeStr)

			if i == m.Cursor {
				s += selectedStyle.Render("> "+line) + "\n"
			} else {
				s += unselectedStyle.Render("  "+line) + "\n"
			}
		}
	}

	// Viewport scroll indicator
	total := len(m.Entries)
	if total > 0 {
		s += "\n" + mutedStyle.Render(fmt.Sprintf("Showing %d-%d of %d items", m.ViewportStart+1, endIdx, total)) + "\n"
	}

	// Footer Help
	help := helpStyle.Render("↑/↓/j/k: navigate • ←/h: parent • →/l/Enter: open/serve • r: refresh • q: quit")
	s += help

	return s
}

func (m Model) viewServing() string {
	var s string

	// Header (1 line + 1 line space = 2 lines)
	header := statusLiveStyle.Render(" SERVING LIVE ") + " " + pathStyle.Render(m.ServedFile.Name)
	s += header + "\n\n"

	if m.Err != nil {
		s += errorBoxStyle.Render("Error: "+m.Err.Error()) + "\n"
	}

	isCompact := m.Height < 22

	// File Info Card
	mimeType := detectMimeType(m.ServedFile.Path)
	var fileCardContent string
	if isCompact {
		fileCardContent = fmt.Sprintf(
			"File: %s | Size: %s | Type: %s",
			pathStyle.Render(m.ServedFile.Name),
			secondaryTextStyle.Render(formatSize(m.ServedFile.Size)),
			warningTextStyle.Render(mimeType),
		)
	} else {
		fileCardContent = fmt.Sprintf(
			"File: %s\nPath: %s\nSize: %s\nType: %s",
			pathStyle.Render(m.ServedFile.Name),
			mutedStyle.Render(m.ServedFile.Path),
			secondaryTextStyle.Render(formatSize(m.ServedFile.Size)),
			warningTextStyle.Render(mimeType),
		)
	}
	fileCardRendered := cardBoxStyle.Render(fileCardContent)
	fileCardLines := lipgloss.Height(fileCardRendered)
	s += fileCardRendered + "\n"

	// URLs Card
	port := 8080
	if m.Server != nil {
		port = m.Server.Port
	}

	localURL := fmt.Sprintf("http://localhost:%d/", port)
	var urlsText string
	if isCompact {
		urlsText = fmt.Sprintf("Local: %s", accentTextStyle.Render(localURL))
		for _, ip := range m.LANIPs {
			urlsText += fmt.Sprintf(" | LAN: http://%s:%d/", ip, port)
		}
	} else {
		urlsText = fmt.Sprintf("Local URL:   %s", accentTextStyle.Render(localURL))
		for _, ip := range m.LANIPs {
			lanURL := fmt.Sprintf("http://%s:%d/", ip, port)
			urlsText += fmt.Sprintf("\nNetwork LAN: %s", secondaryTextStyle.Render(lanURL))
		}
		urlsText += "\n\n" + mutedStyle.Render("Optimized for streaming in mpv, vlc, browser, or curl (Range Requests enabled)")
	}

	urlCardRendered := urlBoxStyle.Render(urlsText)
	urlCardLines := lipgloss.Height(urlCardRendered)
	s += urlCardRendered + "\n"

	// Calculate remaining height for request logs
	usedHeight := 2 + fileCardLines + 1 + urlCardLines + 1 + 1 + 1
	if m.Err != nil {
		usedHeight += 1
	}

	availLogLines := m.Height - usedHeight
	if availLogLines < 1 {
		availLogLines = 1
	}

	// Request Logs Section
	scrollInfo := ""
	if m.LogOffset > 0 {
		scrollInfo = fmt.Sprintf(" (scrolled -%d)", m.LogOffset)
	}
	s += titleStyle.Render(" REQUEST LOGS ") + mutedStyle.Render(scrollInfo) + "\n"

	if len(m.Logs) == 0 {
		s += mutedStyle.Render("Waiting for incoming HTTP requests...") + "\n"
	} else {
		totalLogs := len(m.Logs)
		maxOffset := totalLogs - availLogLines
		if maxOffset < 0 {
			maxOffset = 0
		}
		offset := m.LogOffset
		if offset > maxOffset {
			offset = maxOffset
		}

		endIdx := totalLogs - offset
		if endIdx > totalLogs {
			endIdx = totalLogs
		}
		if endIdx < 0 {
			endIdx = 0
		}

		startIdx := endIdx - availLogLines
		if startIdx < 0 {
			startIdx = 0
		}

		for i := startIdx; i < endIdx; i++ {
			if i >= len(m.Logs) {
				break
			}
			logEvt := m.Logs[i]
			statusStr := fmt.Sprintf("%d", logEvt.Status)
			if logEvt.Status >= 200 && logEvt.Status < 300 {
				statusStr = secondaryTextStyle.Render(statusStr)
			} else if logEvt.Status >= 300 && logEvt.Status < 400 {
				statusStr = warningTextStyle.Render(statusStr)
			} else {
				statusStr = errorBoxStyle.Render(statusStr)
			}

			timeStr := logEvt.Timestamp.Format("15:04:05")
			line := fmt.Sprintf("[%s] %s | %s %s | %s | %s in %v",
				mutedStyle.Render(timeStr),
				accentTextStyle.Render(logEvt.ClientIP),
				warningTextStyle.Render(logEvt.Method),
				logEvt.Path,
				statusStr,
				formatSize(logEvt.Bytes),
				logEvt.Duration.Truncate(time.Millisecond),
			)
			s += line + "\n"
		}
	}

	// Footer Help
	help := helpStyle.Render("↑/↓: scroll logs • Esc: stop server • q: quit app")
	s += help

	return s
}

func formatSize(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

var customMimeTypes = map[string]string{
	".mkv":  "video/x-matroska",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".wmv":  "video/x-ms-wmv",
	".flv":  "video/x-flv",
	".m4v":  "video/x-m4v",
	".mp3":  "audio/mpeg",
	".flac": "audio/flac",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".m4a":  "audio/mp4",
	".aac":  "audio/aac",
	".webp": "image/webp",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".pdf":  "application/pdf",
	".json": "application/json",
	".xml":  "application/xml",
	".zip":  "application/zip",
	".tar":  "application/x-tar",
	".gz":   "application/gzip",
	".txt":  "text/plain; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "application/javascript; charset=utf-8",
	".go":   "text/plain; charset=utf-8",
	".py":   "text/plain; charset=utf-8",
	".md":   "text/markdown; charset=utf-8",
}

func detectMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if t, ok := customMimeTypes[ext]; ok {
		return t
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}
