package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

// DrivesVirtualPath represents the top-level "This PC" drives view.
const DrivesVirtualPath = "This PC"

// Entry represents a file or directory item in the directory listing.
type Entry struct {
	Name     string
	Path     string
	IsDir    bool
	IsParent bool
	Size     int64
	ModTime  time.Time
}

// caseInsensitiveLess compares two strings case-insensitively without heap allocations.
func caseInsensitiveLess(a, b string) bool {
	la, lb := len(a), len(b)
	minLen := la
	if lb < minLen {
		minLen = lb
	}
	for i := 0; i < minLen; i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return ca < cb
		}
	}
	if la == lb {
		return a < b
	}
	return la < lb
}

// ReadDirectory reads the contents of the given directory path,
// prepends a parent directory ("..") entry if applicable,
// and sorts directories before files (alphabetically).
func ReadDirectory(dirPath string) ([]Entry, string, error) {
	if dirPath == DrivesVirtualPath {
		if runtime.GOOS != "windows" {
			dirPath = "/"
		} else {
			drives := GetLogicalDrives()
			var entries []Entry
			for _, d := range drives {
				entries = append(entries, Entry{
					Name:  d,
					Path:  d,
					IsDir: true,
				})
			}
			return entries, DrivesVirtualPath, nil
		}
	}

	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve absolute path for %s: %w", dirPath, err)
	}

	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, absPath, fmt.Errorf("failed to read directory %s: %w", absPath, err)
	}

	var dirs []Entry
	var files []Entry

	// Check parent directory
	parentPath := filepath.Dir(absPath)
	if parentPath != absPath {
		dirs = append(dirs, Entry{
			Name:     "..",
			Path:     parentPath,
			IsDir:    true,
			IsParent: true,
		})
	} else if runtime.GOOS == "windows" {
		// At drive root (e.g. C:\) on Windows -> parent goes to "This PC"
		dirs = append(dirs, Entry{
			Name:     "..",
			Path:     DrivesVirtualPath,
			IsDir:    true,
			IsParent: true,
		})
	}

	for _, de := range dirEntries {
		fullPath := filepath.Join(absPath, de.Name())
		isDir := de.IsDir()
		var size int64
		var modTime time.Time

		// Performance: only query info/stat for files, avoiding redundant stat syscalls on directories
		if !isDir {
			if info, err := de.Info(); err == nil {
				size = info.Size()
				modTime = info.ModTime()
			}
		}

		entry := Entry{
			Name:     de.Name(),
			Path:     fullPath,
			IsDir:    isDir,
			IsParent: false,
			Size:     size,
			ModTime:  modTime,
		}

		if isDir {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	// Sort directories (excluding '..' which is already at index 0)
	sortStart := 0
	if len(dirs) > 0 && dirs[0].IsParent {
		sortStart = 1
	}
	if sortStart < len(dirs) {
		sort.Slice(dirs[sortStart:], func(i, j int) bool {
			return caseInsensitiveLess(dirs[sortStart+i].Name, dirs[sortStart+j].Name)
		})
	}

	// Sort files alphabetically
	sort.Slice(files, func(i, j int) bool {
		return caseInsensitiveLess(files[i].Name, files[j].Name)
	})

	entries := append(dirs, files...)
	return entries, absPath, nil
}
