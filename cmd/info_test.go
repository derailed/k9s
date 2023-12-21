// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_getScreenDumpDirForInfo(t *testing.T) {
	tests := map[string]struct {
		k9sConfigFile         string
		expectedScreenDumpDir string
	}{
		"withK9sConfigFile": {
			k9sConfigFile:         "testdata/k9s.yaml",
			expectedScreenDumpDir: "/tmp",
		},
		"withEmptyK9sConfigFile": {
			k9sConfigFile:         "",
			expectedScreenDumpDir: config.AppDumpsDir,
		},
		"withInvalidK9sConfigFilePath": {
			k9sConfigFile:         "invalid",
			expectedScreenDumpDir: config.AppDumpsDir,
		},
		"withScreenDumpDirEmptyInK9sConfigFile": {
			k9sConfigFile:         "testdata/k9s1.yaml",
			expectedScreenDumpDir: config.AppDumpsDir,
		},
	}
	for k := range tests {
		u := tests[k]
		t.Run(k, func(t *testing.T) {
			initK9sConfigFile := config.AppConfigFile
			config.AppConfigFile = u.k9sConfigFile

			assert.Equal(t, u.expectedScreenDumpDir, getScreenDumpDirForInfo())

			config.AppConfigFile = initK9sConfigFile
		})
	}
}
