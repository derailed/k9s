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

const (
	envPFAddress          = "K9S_DEFAULT_PF_ADDRESS"
	envFGNodeShell        = "K9S_FEATURE_GATE_NODE_SHELL"
	defaultPortFwdAddress = "localhost"
)

var invalidPathCharsRX = regexp.MustCompile(`[:/]+`)

// SanitizeContextSubpath ensure cluster/context produces a valid path.
func SanitizeContextSubpath(cluster, context string) string {
	return filepath.Join(SanitizeFileName(cluster), SanitizeFileName(context))
}

// SanitizeFileName ensure file spec is valid.
func SanitizeFileName(name string) string {
	return invalidPathCharsRX.ReplaceAllString(name, "-")
}

func defaultPFAddress() string {
	if a := os.Getenv(envPFAddress); a != "" {
		return a
	}

	return defaultPortFwdAddress
}

func defaultFGNodeShell() bool {
	if a := os.Getenv(envFGNodeShell); a != "" {
		return a == "true"
	}

	return false
}

// InList check if string is in a collection of strings.
func InList(ll []string, n string) bool {
	for _, l := range ll {
		if l == n {
			return true
		}
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
