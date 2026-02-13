//go:build !windows

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

	tmp := exe + ".new"
	if err := os.WriteFile(tmp, newBinary, 0755); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Run: sudo switcher update")
		}
		return fmt.Errorf("writing new binary: %w", err)
	}

	if err := os.Rename(tmp, exe); err != nil {
		os.Remove(tmp)
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Run: sudo switcher update")
		}
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}
