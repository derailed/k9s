// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/stretchr/testify/assert"
)

func TestBuildContextCommand(t *testing.T) {
	tests := []struct {
		name            string
		contextName     string
		expectedContext string
	}{
		{
			name:            "context without whitespace",
			contextName:     "my-cluster",
			expectedContext: "my-cluster",
		},
		{
			name:            "context with whitespace",
			contextName:     "ctx1 (with whitespace)",
			expectedContext: "ctx1 (with whitespace)",
		},
		{
			name:            "context with multiple spaces",
			contextName:     "my cluster name",
			expectedContext: "my cluster name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the command string
			cmdString := buildContextCommand(tt.contextName)

			// Verify the interpreter can extract the correct context name from it
			interp := cmd.NewInterpreter(cmdString)
			gotContext, ok := interp.ContextArg()

			assert.True(t, ok, "should recognize as context command")
			assert.Equal(t, tt.expectedContext, gotContext,
				"buildContextCommand must generate a format that preserves context names with whitespace")
		})
	}
}
