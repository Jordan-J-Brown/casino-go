package testingx

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrFailedToFindProjectRoot = errors.New("failed to find project root")
)

func ProjectRoot(dir string) (string, error) {
	vol := filepath.VolumeName(dir)
	if vol == "" {
		vol = "/"
	}

	for dir != vol {
		_, pathErr := os.Stat(filepath.Join(dir, "go.mod"))
		if pathErr == nil {
			return dir, nil
		}

		dir = filepath.Dir(dir)
	}

	return "", ErrFailedToFindProjectRoot
}
