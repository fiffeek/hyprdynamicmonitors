package utils

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const XDGRuntimeDir = "XDG_RUNTIME_DIR"

func GetXDGRuntimeDir() (string, error) {
	xdgRuntimeDir := os.Getenv(XDGRuntimeDir)

	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error while getting the current user: %w", err)
	}

	if xdgRuntimeDir == "" {
		return "", errors.New("XDG_RUNTIME_DIR environment variable not set")
	}

	if xdgRuntimeDir == "" {
		user := u.Uid
		xdgRuntimeDir = filepath.Join("/run/user", user)
	}

	return xdgRuntimeDir, nil
}
