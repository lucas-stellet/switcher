//go:build windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

func replaceBinary(newBinary []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating current binary: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	old := exe + ".old"

	// Rename running binary out of the way (Windows locks running executables)
	if err := os.Rename(exe, old); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Try running as administrator")
		}
		return fmt.Errorf("renaming current binary: %w", err)
	}

	if err := os.WriteFile(exe, newBinary, 0755); err != nil {
		// Try to restore the old binary
		os.Rename(old, exe)
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Try running as administrator")
		}
		return fmt.Errorf("writing new binary: %w", err)
	}

	// Best-effort cleanup of old binary
	os.Remove(old)

	return nil
}
