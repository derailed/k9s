// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/assert"
)

func TestHelperInList(t *testing.T) {
	uu := []struct {
		item     string
		list     []string
		expected bool
	}{
		{"a", []string{}, false},
		{"", []string{}, false},
		{"", []string{""}, true},
		{"a", []string{"a", "b", "c", "d"}, true},
		{"z", []string{"a", "b", "c", "d"}, false},
	}

	for _, u := range uu {
		assert.Equal(t, u.expected, data.InList(u.list, u.item))
	}
}

func TestEnsureDirPathNone(t *testing.T) {
	var mod os.FileMode = 0744
	dir := filepath.Join("/tmp", "fred")
	os.Remove(dir)

	path := filepath.Join(dir, "duh.yaml")
	assert.NoError(t, data.EnsureDirPath(path, mod))

	p, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.Equal(t, "drwxr--r--", p.Mode().String())
}

func TestEnsureDirPathNoOpt(t *testing.T) {
	var mod os.FileMode = 0744
	dir := filepath.Join("/tmp", "blee")
	os.Remove(dir)
	assert.NoError(t, os.Mkdir(dir, mod))

	path := filepath.Join(dir, "duh.yaml")
	assert.NoError(t, data.EnsureDirPath(path, mod))

	p, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.Equal(t, "drwxr--r--", p.Mode().String())
}
