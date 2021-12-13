package cmd

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getScreenDumpDirForInfo(t *testing.T) {
	tests := []struct {
		name                  string
		k9sConfigFile         string
		expectedScreenDumpDir string
	}{
		{
			name:                  "withK9sConfigFile",
			k9sConfigFile:         "testdata/k9s.yml",
			expectedScreenDumpDir: "/tmp",
		},
		{
			name:                  "withEmptyK9sConfigFile",
			k9sConfigFile:         "",
			expectedScreenDumpDir: config.K9sDefaultScreenDumpDir,
		},
		{
			name:                  "withInvalidK9sConfigFilePath",
			k9sConfigFile:         "invalid",
			expectedScreenDumpDir: config.K9sDefaultScreenDumpDir,
		},
		{
			name:                  "withScreenDumpDirEmptyInK9sConfigFile",
			k9sConfigFile:         "testdata/k9s1.yml",
			expectedScreenDumpDir: config.K9sDefaultScreenDumpDir,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initK9sConfigFile := config.K9sConfigFile

			config.K9sConfigFile = tt.k9sConfigFile

			assert.Equal(t, tt.expectedScreenDumpDir, getScreenDumpDirForInfo())

			config.K9sConfigFile = initK9sConfigFile
		})
	}
}
