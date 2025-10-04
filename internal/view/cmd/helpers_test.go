// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func Test_toLabels(t *testing.T) {
	uu := map[string]struct {
		s  string
		ll map[string]string
	}{
		"empty": {},
		"toast": {
			s: "=",
		},
		"toast-1": {
			s: "=,",
		},
		"toast-2": {
			s: ",",
		},
		"toast-3": {
			s: ",=",
		},
		"simple": {
			s:  "a=b",
			ll: map[string]string{"a": "b"},
		},
		"multi": {
			s:  "a=b,c=d",
			ll: map[string]string{"a": "b", "c": "d"},
		},
		"multi-toast1": {
			s:  "a=,c=d",
			ll: map[string]string{"c": "d"},
		},
		"multi-toast2": {
			s:  "a=b,=d",
			ll: map[string]string{"a": "b"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.ll, ToLabels(u.s))
		})
	}
}

func TestSuggestSubCommand(t *testing.T) {
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
		{Command: "ctx", Suggestions: []string{" develop", " pre", " prod", " test"}},
		{Command: "ctx ", Suggestions: []string{"develop", "pre", "prod", "test"}},
		{Command: "context   d", Suggestions: []string{"evelop"}},
		{Command: "contexts   t", Suggestions: []string{"est"}},
		{Command: "po ", Suggestions: nil},
		{Command: "po  x", Suggestions: nil},
		{Command: "po k", Suggestions: []string{"ube-public", "ube-system"}},
		{Command: "po  kube-", Suggestions: []string{"public", "system"}},
	}

	for _, tt := range tests {
		got := SuggestSubCommand(tt.Command, namespaceNames, contextNames)
		assert.Equal(t, tt.Suggestions, got)
	}
}
