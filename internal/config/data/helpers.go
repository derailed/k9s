// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"os"
	"path/filepath"
)

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
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, mod); err != nil {
			return err
		}
	}

	return nil
}
