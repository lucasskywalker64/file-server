//go:build windows

package fs

import (
	"os"

	"golang.org/x/sys/windows"
)

// GetLogicalDrives returns a list of available drive roots (e.g. ["C:\\", "D:\\"] on Windows).
func GetLogicalDrives() []string {
	var drives []string
	mask, err := windows.GetLogicalDrives()
	if err == nil && mask != 0 {
		for i := uint32(0); i < 26; i++ {
			if mask&(1<<i) != 0 {
				driveLetter := string(rune('A'+i)) + ":\\"
				drives = append(drives, driveLetter)
			}
		}
		return drives
	}

	// Fallback scan A-Z on Windows
	for _, b := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drive := string(b) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}

	if len(drives) == 0 {
		drives = append(drives, "C:\\")
	}

	return drives
}
