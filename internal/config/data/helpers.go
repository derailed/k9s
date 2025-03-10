// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

const envFGNodeShell = "K9S_FEATURE_GATE_NODE_SHELL"

var invalidPathCharsRX = regexp.MustCompile(`[:/]+`)

// SanitizeContextSubpath ensure cluster/context produces a valid path.
func SanitizeContextSubpath(cluster, context string) string {
	return filepath.Join(SanitizeFileName(cluster), SanitizeFileName(context))
}

// SanitizeFileName ensure file spec is valid.
func SanitizeFileName(name string) string {
	return invalidPathCharsRX.ReplaceAllString(name, "-")
}

func defaultFGNodeShell() bool {
	if a := os.Getenv(envFGNodeShell); a != "" {
		return a == "true"
	}

	return false
}

// EnsureDirPath ensures a directory exist from the given path.
func EnsureDirPath(path string, mod os.FileMode) error {
	return EnsureFullPath(filepath.Dir(path), mod)
}

// EnsureFullPath ensures a directory exist from the given path.
func EnsureFullPath(path string, mod os.FileMode) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if err = os.MkdirAll(path, mod); err != nil {
			return err
		}
	}

	return nil
}
