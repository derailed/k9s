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

func TestSuggestSubCommandPrefix(t *testing.T) {
	namespaceNames := map[string]struct{}{
		"kube-system":   {},
		"kube-public":   {},
		"default":       {},
		"nginx-ingress": {},
	}
	contextNames := []string{"develop", "test", "pre", "prod"}

	tests := []struct {
		Command     string
		Suggestions map[string]float64
	}{
		{Command: "q", Suggestions: nil},
		{Command: "xray  dp", Suggestions: nil},
		{Command: "help  k", Suggestions: nil},
		{Command: "ctx p", Suggestions: map[string]float64{"pre": 0, "prod": 0}},
		{Command: "ctx   p", Suggestions: map[string]float64{"pre": 0, "prod": 0}},
		{Command: "ctx pr", Suggestions: map[string]float64{"pre": 0, "prod": 0}},
		{Command: "ctx", Suggestions: map[string]float64{" develop": 0, " pre": 0, " prod": 0, " test": 0}},
		{Command: "ctx ", Suggestions: map[string]float64{"develop": 0, "pre": 0, "prod": 0, "test": 0}},
		{Command: "context   d", Suggestions: map[string]float64{"develop": 0}},
		{Command: "contexts   t", Suggestions: map[string]float64{"test": 0}},
		{Command: "po ", Suggestions: nil},
		{Command: "po  x", Suggestions: nil},
		{Command: "po k", Suggestions: map[string]float64{"kube-public": 0, "kube-system": 0}},
		{Command: "po  kube-", Suggestions: map[string]float64{"kube-public": 0, "kube-system": 0}},
	}

	for _, tt := range tests {
		got := SuggestSubCommand("PREFIX", tt.Command, namespaceNames, contextNames)
		assert.Equal(t, tt.Suggestions, got)
	}
}

func TestSuggestSubCommandLongestPrefix(t *testing.T) {
	namespaceNames := map[string]struct{}{
		"kube-system":   {},
		"kube-public":   {},
		"default":       {},
		"nginx-ingress": {},
	}
	contextNames := []string{"develop", "test", "pre", "prod"}

	tests := []struct {
		Command     string
		Suggestions map[string]float64
	}{
		{Command: "q", Suggestions: nil},
		{Command: "xray  dp", Suggestions: nil},
		{Command: "help  k", Suggestions: nil},
		{Command: "ctx p", Suggestions: map[string]float64{"pre": 1.0 / 3.0, "prod": 1.0 / 4.0}},
		{Command: "ctx   p", Suggestions: map[string]float64{"pre": 1.0 / 3.0, "prod": 1.0 / 4.0}},
		{Command: "ctx pr", Suggestions: map[string]float64{"pre": 2.0 / 3.0, "prod": 2.0 / 4.0}},
		{Command: "ctx", Suggestions: map[string]float64{" develop": 0, " pre": 0, " prod": 0, " test": 0}},
		{Command: "ctx ", Suggestions: map[string]float64{"develop": 0, "pre": 0, "prod": 0, "test": 0}},
		{Command: "context   d", Suggestions: map[string]float64{"develop": 1.0 / 7.0}},
		{Command: "contexts   t", Suggestions: map[string]float64{"test": 1.0 / 4.0}},
		{Command: "po ", Suggestions: nil},
		{Command: "po  x", Suggestions: nil},
		{Command: "po k", Suggestions: map[string]float64{"kube-public": 1.0 / 11.0, "kube-system": 1.0 / 11.0}},
		{Command: "po  kube-", Suggestions: map[string]float64{"kube-public": 5.0 / 11.0, "kube-system": 5.0 / 11.0}},
	}

	for _, tt := range tests {
		got := SuggestSubCommand("LONGEST_PREFIX", tt.Command, namespaceNames, contextNames)
		assert.Equal(t, tt.Suggestions, got)
	}
}

func TestSuggestSubCommandLongestSubstring(t *testing.T) {
	namespaceNames := map[string]struct{}{
		"kube-system":   {},
		"kube-public":   {},
		"default":       {},
		"nginx-ingress": {},
	}
	contextNames := []string{"develop", "test", "pre", "prod"}

	tests := []struct {
		Command     string
		Suggestions map[string]float64
	}{
		{Command: "q", Suggestions: nil},
		{Command: "xray  dp", Suggestions: nil},
		{Command: "help  k", Suggestions: nil},
		{Command: "ctx p", Suggestions: map[string]float64{"develop": 1.0 / 7.0, "pre": 1.0 / 3.0, "prod": 1.0 / 4.0}},
		{Command: "ctx   p", Suggestions: map[string]float64{"develop": 1.0 / 7.0, "pre": 1.0 / 3.0, "prod": 1.0 / 4.0}},
		{Command: "ctx pr", Suggestions: map[string]float64{"pre": 2.0 / 3.0, "prod": 2.0 / 4.0}},
		{Command: "ctx", Suggestions: map[string]float64{" develop": 0, " pre": 0, " prod": 0, " test": 0}},
		{Command: "ctx ", Suggestions: map[string]float64{"develop": 0, "pre": 0, "prod": 0, "test": 0}},
		{Command: "context   d", Suggestions: map[string]float64{"develop": 1.0 / 7.0, "prod": 1.0 / 4.0}},
		{Command: "contexts   t", Suggestions: map[string]float64{"test": 1.0 / 4.0}},
		{Command: "po ", Suggestions: nil},
		{Command: "po  x", Suggestions: map[string]float64{"nginx-ingress": 1.0 / 13.0}},
		{Command: "po k", Suggestions: map[string]float64{"kube-public": 1.0 / 11.0, "kube-system": 1.0 / 11.0}},
		{Command: "po  kube-", Suggestions: map[string]float64{"kube-public": 5.0 / 11.0, "kube-system": 5.0 / 11.0}},
	}

	for _, tt := range tests {
		got := SuggestSubCommand("LONGEST_SUBSTRING", tt.Command, namespaceNames, contextNames)
		assert.Equal(t, tt.Suggestions, got)
	}
}
