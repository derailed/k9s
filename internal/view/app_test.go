// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAppNew(t *testing.T) {
	a := NewApp(config.NewConfig(ks{}))
	_ = a.Init("blee", 10)

	assert.Equal(t, 11, len(a.GetActions()))
}

func Test_suggestSubCommand(t *testing.T) {
	namespaceNames := []string{"kube-system", "kube-public", "default", "nginx-ingress"}
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
		{Command: "po k", Suggestions: []string{"ube-system", "ube-public"}},
		{Command: "po  kube-", Suggestions: []string{"system", "public"}},
	}

	for _, tt := range tests {
		got := suggestSubCommand(tt.Command, namespaceNames, contextNames)
		assert.Equal(t, tt.Suggestions, got)
	}
}
