// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_suggestSubCommand(t *testing.T) {
	namespaceNames := map[string]struct{}{
		"kube-system":   {},
		"kube-public":   {},
		"default":       {},
		"nginx-ingress": {},
	}
	contextNames := []string{"develop", "test", "pre", "prod"}

	tests := []struct {
		Command     string
		Suggestions []string
	}{
		{Command: "q", Suggestions: nil},
		{Command: "xray  dp", Suggestions: nil},
		{Command: "help  k", Suggestions: nil},
		{Command: "ctx p", Suggestions: []string{"re", "rod"}},
		{Command: "ctx   p", Suggestions: []string{"re", "rod"}},
		{Command: "ctx pr", Suggestions: []string{"e", "od"}},
		{Command: "context   d", Suggestions: []string{"evelop"}},
		{Command: "contexts   t", Suggestions: []string{"est"}},
		{Command: "po ", Suggestions: nil},
		{Command: "po  x", Suggestions: nil},
		{Command: "po k", Suggestions: []string{"ube-public", "ube-system"}},
		{Command: "po  kube-", Suggestions: []string{"public", "system"}},
	}

	for _, tt := range tests {
		got := suggestSubCommand(tt.Command, namespaceNames, contextNames)
		assert.Equal(t, tt.Suggestions, got)
	}
}
