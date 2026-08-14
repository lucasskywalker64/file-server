package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create test structure:
	// - dirB/
	// - dirA/
	// - fileB.txt
	// - fileA.txt
	err := os.Mkdir(filepath.Join(tempDir, "dirB"), 0755)
	if err != nil {
		t.Fatalf("failed to create dirB: %v", err)
	}
	err = os.Mkdir(filepath.Join(tempDir, "dirA"), 0755)
	if err != nil {
		t.Fatalf("failed to create dirA: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "fileB.txt"), []byte("file B content"), 0644)
	if err != nil {
		t.Fatalf("failed to create fileB: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "fileA.txt"), []byte("file A content"), 0644)
	if err != nil {
		t.Fatalf("failed to create fileA: %v", err)
	}

	entries, absPath, err := ReadDirectory(tempDir)
	if err != nil {
		t.Fatalf("ReadDirectory returned error: %v", err)
	}

	if absPath == "" {
		t.Errorf("expected non-empty absPath")
	}

	var names []string
	for _, e := range entries {
		names = append(names, e.Name)
	}

	if len(names) < 5 {
		t.Fatalf("expected at least 5 entries, got %d: %v", len(names), names)
	}

	if names[0] != ".." {
		t.Errorf("expected first entry to be '..', got '%s'", names[0])
	}
	if names[1] != "dirA" || names[2] != "dirB" {
		t.Errorf("expected dirs sorted ['dirA', 'dirB'], got ['%s', '%s']", names[1], names[2])
	}
	if names[3] != "fileA.txt" || names[4] != "fileB.txt" {
		t.Errorf("expected files sorted ['fileA.txt', 'fileB.txt'], got ['%s', '%s']", names[3], names[4])
	}
}

func TestReadDirectory_DrivesVirtualPath(t *testing.T) {
	entries, path, err := ReadDirectory(DrivesVirtualPath)
	if err != nil {
		t.Fatalf("ReadDirectory(DrivesVirtualPath) failed: %v", err)
	}
	if path != DrivesVirtualPath {
		t.Errorf("expected path '%s', got '%s'", DrivesVirtualPath, path)
	}
	if len(entries) == 0 {
		t.Errorf("expected at least one drive entry")
	}

	foundC := false
	for _, e := range entries {
		if e.Name == "C:\\" {
			foundC = true
			break
		}
	}
	if !foundC {
		t.Errorf("expected 'C:\\' drive in entries, got %v", entries)
	}
}

func TestCaseInsensitiveLess(t *testing.T) {
	tests := []struct {
		a, b     string
		expected bool
	}{
		{"apple", "Banana", true},
		{"Banana", "apple", false},
		{"File1.txt", "file2.txt", true},
		{"test", "test", false},
		{"a", "ab", true},
		{"ab", "a", false},
	}

	for _, tc := range tests {
		got := caseInsensitiveLess(tc.a, tc.b)
		if got != tc.expected {
			t.Errorf("caseInsensitiveLess(%q, %q) = %v; expected %v", tc.a, tc.b, got, tc.expected)
		}
	}
}
