//go:build !windows

package fs

// GetLogicalDrives returns a list of available drive roots (["/"] on non-Windows platforms).
func GetLogicalDrives() []string {
	return []string{"/"}
}
