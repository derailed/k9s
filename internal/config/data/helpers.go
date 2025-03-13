// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
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

// WriteYAML writes a yaml file to bytes.
func WriteYAML(content any) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	ec := yaml.NewEncoder(buff)
	ec.SetIndent(2)

	if err := ec.Encode(content); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// SaveYAML writes a yaml file to disk.
func SaveYAML(path string, content any) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, DefaultFileMod)
	if err != nil {
		return err
	}
	ec := yaml.NewEncoder(f)
	ec.SetIndent(2)

	return ec.Encode(content)
}
