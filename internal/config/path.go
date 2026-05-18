package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultPath returns the platform default icuvisor config path.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config directory: %w", err)
	}
	return filepath.Join(dir, "icuvisor", "config.json"), nil
}
