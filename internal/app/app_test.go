package app

import (
	"os"
	"path/filepath"
	"testing"

	"file-server/internal/fs"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAppModel_NavigationAndServing(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "video.mp4")
	err := os.WriteFile(testFile, []byte("fake mp4 content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	model, err := NewModel(tempDir, 19090)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}

	if model.State != StateNavigating {
		t.Errorf("expected initial state StateNavigating, got %v", model.State)
	}

	// Move cursor down to select video.mp4 (index 0 is "..")
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updatedModel.(Model)

	if m.Cursor != 1 {
		t.Errorf("expected cursor at 1 after KeyDown, got %d", m.Cursor)
	}

	// Press Enter to start serving file
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(Model)

	if m.State != StateServing {
		t.Errorf("expected state StateServing after Enter on file, got %v", m.State)
	}
	if m.Server == nil {
		t.Fatalf("expected active server instance")
	}
	if cmd == nil {
		t.Errorf("expected tea.Cmd for log listener")
	}

	// Clean up server
	defer m.Server.Stop()

	// Press Esc to stop server and return to navigating
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(Model)

	if m.State != StateNavigating {
		t.Errorf("expected state StateNavigating after Esc, got %v", m.State)
	}
	if m.Server != nil {
		t.Errorf("expected server to be nil after stopping")
	}
}

func TestAppModel_QuitKeyInServingState(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.mp4")
	_ = os.WriteFile(testFile, []byte("data"), 0644)

	model, err := NewModel(tempDir, 19092)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}

	// Select file and press Enter
	model.Cursor = 1
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	if m.State != StateServing {
		t.Fatalf("expected state StateServing")
	}

	// Press 'q' while in serving state
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected tea.Quit command when pressing 'q' in serving state")
	}
}

func TestAppModel_NavigateToDriveRoot(t *testing.T) {
	model, err := NewModel("C:\\", 19091)
	if err != nil {
		t.Fatalf("NewModel at C:\\ failed: %v", err)
	}

	// First item at C:\ should be ".." with path "This PC"
	if len(model.Entries) == 0 || model.Entries[0].Name != ".." {
		t.Fatalf("expected first entry at C:\\ to be '..'")
	}
	if model.Entries[0].Path != fs.DrivesVirtualPath {
		t.Errorf("expected parent path of C:\\ to be '%s', got '%s'", fs.DrivesVirtualPath, model.Entries[0].Path)
	}

	// Press Enter on ".." to go to "This PC"
	model.Cursor = 0
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	if m.Cwd != fs.DrivesVirtualPath {
		t.Errorf("expected Cwd to be '%s', got '%s'", fs.DrivesVirtualPath, m.Cwd)
	}

	// Should contain drive list
	if len(m.Entries) == 0 {
		t.Errorf("expected drive entries in 'This PC'")
	}
}

func TestAppModel_VimAndArrowNavigation(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subfolder")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644)

	model, err := NewModel(tempDir, 19093)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}

	// Navigate with 'j' (down) to select subfolder
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updated.(Model)
	if m.Cursor != 1 {
		t.Errorf("expected cursor at 1 after 'j', got %d", m.Cursor)
	}

	// Enter subfolder using 'l'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(Model)
	if m.Cwd != subDir {
		t.Errorf("expected Cwd %s after 'l', got %s", subDir, m.Cwd)
	}

	// Go back to parent using 'h'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(Model)
	if m.Cwd != tempDir {
		t.Errorf("expected Cwd %s after 'h', got %s", tempDir, m.Cwd)
	}

	// Refresh with 'r'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(Model)
	if len(m.Entries) == 0 {
		t.Errorf("expected non-empty entries after 'r'")
	}
}

func TestAppModel_LogScrollingBounds(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("data"), 0644)

	model, err := NewModel(tempDir, 19094)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}

	// Serve file
	model.Cursor = 1
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	defer m.Server.Stop()

	// Simulate adding logs
	for i := 0; i < 20; i++ {
		updated, _ = m.Update(LogEventMsg{})
		m = updated.(Model)
	}

	// Scroll up with 'k'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.LogOffset != 1 {
		t.Errorf("expected LogOffset 1, got %d", m.LogOffset)
	}

	// Jump to top with 'g'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updated.(Model)
	if m.LogOffset != 19 {
		t.Errorf("expected LogOffset 19 at top, got %d", m.LogOffset)
	}

	// Jump to bottom with 'G'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = updated.(Model)
	if m.LogOffset != 0 {
		t.Errorf("expected LogOffset 0 at bottom, got %d", m.LogOffset)
	}

	// Render view serving
	rendered := m.View()
	if len(rendered) == 0 {
		t.Errorf("expected non-empty rendered view")
	}
}

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"movie.mkv", "video/x-matroska"},
		{"clip.mp4", "video/mp4"},
		{"song.flac", "audio/flac"},
		{"data.json", "application/json"},
		{"doc.pdf", "application/pdf"},
		{"unknown.xyz123", "application/octet-stream"},
	}

	for _, tc := range tests {
		got := detectMimeType(tc.filename)
		if got != tc.expected {
			t.Errorf("detectMimeType(%q) = %q; expected %q", tc.filename, got, tc.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{-10, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range tests {
		got := formatSize(tc.size)
		if got != tc.expected {
			t.Errorf("formatSize(%d) = %q; expected %q", tc.size, got, tc.expected)
		}
	}
}
