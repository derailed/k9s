// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

//go:build unix

package data_test

import (
	"path/filepath"
	"syscall"
	"testing"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/require"
)

// TestSaveYAMLNoFDLeak lowers the fd limit so a leaked fd per call
// would fail with "too many open files" within a few dozen iterations.
func TestSaveYAMLNoFDLeak(t *testing.T) {
	var cur syscall.Rlimit
	require.NoError(t, syscall.Getrlimit(syscall.RLIMIT_NOFILE, &cur))
	require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{Cur: 64, Max: cur.Max}))
	defer func() {
		_ = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &cur)
	}()

	path := filepath.Join(t.TempDir(), "config.yaml")
	for i := 1; i <= 200; i++ {
		require.NoErrorf(t, data.SaveYAML(path, map[string]int{"iteration": i}),
			"SaveYAML leaked file descriptors: failed at iteration %d", i)
	}
}
