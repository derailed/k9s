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
			k9sConfigFile:         "testdata/k9s.yml",
			expectedScreenDumpDir: "/tmp",
		},
		"withEmptyK9sConfigFile": {
			k9sConfigFile:         "",
			expectedScreenDumpDir: config.K9sDefaultScreenDumpDir,
		},
		"withInvalidK9sConfigFilePath": {
			k9sConfigFile:         "invalid",
			expectedScreenDumpDir: config.K9sDefaultScreenDumpDir,
		},
		"withScreenDumpDirEmptyInK9sConfigFile": {
			k9sConfigFile:         "testdata/k9s1.yml",
			expectedScreenDumpDir: config.K9sDefaultScreenDumpDir,
		},
	}
	for k := range tests {
		u := tests[k]
		t.Run(k, func(t *testing.T) {
			initK9sConfigFile := config.K9sConfigFile

			config.K9sConfigFile = u.k9sConfigFile

			assert.Equal(t, u.expectedScreenDumpDir, getScreenDumpDirForInfo())

			config.K9sConfigFile = initK9sConfigFile
		})
	}
}
