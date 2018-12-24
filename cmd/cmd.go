package cmd

import (
	"os"
	"path/filepath"
)

func getRealDir(dir string) (string, error) {
	if dir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return dir, nil
	}

	if !filepath.IsAbs(dir) {
		dir, err := filepath.Abs(dir)
		if err != nil {
			return "", err
		}
		return dir, nil
	}

	return dir, nil
}
