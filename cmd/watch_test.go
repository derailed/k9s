// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseSelection(t *testing.T) {
	tests := map[string]struct {
		input   string
		n       int
		want    []int
		wantErr bool
	}{
		"simple":        {input: "1,3,5", n: 5, want: []int{0, 2, 4}},
		"spaces":        {input: " 1 , 2 ", n: 5, want: []int{0, 1}},
		"single":        {input: "2", n: 5, want: []int{1}},
		"dedupe":        {input: "1,1,2", n: 5, want: []int{0, 1}},
		"emptyTokens":   {input: "1,,3", n: 5, want: []int{0, 2}},
		"empty":         {input: "", n: 5, want: nil},
		"blank":         {input: "   ", n: 5, want: nil},
		"notANumber":    {input: "1,x", n: 5, wantErr: true},
		"zero":          {input: "0", n: 5, wantErr: true},
		"outOfRange":    {input: "6", n: 5, wantErr: true},
		"negative":      {input: "-1", n: 5, wantErr: true},
		"orderPreserve": {input: "3,1", n: 5, want: []int{2, 0}},
	}

	for k := range tests {
		u := tests[k]
		t.Run(k, func(t *testing.T) {
			got, err := parseSelection(u.input, u.n)
			if u.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, u.want, got)
		})
	}
}

func TestWorkloadEntry(t *testing.T) {
	tests := map[string]struct {
		gvr  string
		ns   string
		want string
	}{
		"noNamespace":     {gvr: "v1/pods", ns: "", want: "v1/pods"},
		"withNamespace":   {gvr: "apps/v1/deployments", ns: "kube-system", want: "apps/v1/deployments kube-system"},
		"allFollowsView":  {gvr: "v1/pods", ns: "all", want: "v1/pods"},
		"trimsWhitespace": {gvr: " v1/pods ", ns: " demo ", want: "v1/pods demo"},
	}

	for k := range tests {
		u := tests[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.want, workloadEntry(u.gvr, u.ns))
		})
	}
}

func TestResolveNamespace(t *testing.T) {
	nss := []string{"default", "demo", "kube-system"}

	assert.Equal(t, "", resolveNamespace("", nss))
	assert.Equal(t, "default", resolveNamespace("1", nss))
	assert.Equal(t, "kube-system", resolveNamespace("3", nss))
	assert.Equal(t, "custom", resolveNamespace("custom", nss))
	// Out of range number falls through to a literal name.
	assert.Equal(t, "9", resolveNamespace("9", nss))
}

func TestMergeWorkloadsFresh(t *testing.T) {
	out, err := mergeWorkloads(nil, []string{"v1/pods", "apps/v1/deployments demo"})
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(out, &doc))
	wl, ok := doc["workloads"].(map[string]any)
	require.True(t, ok)
	def, ok := wl["default"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"v1/pods", "apps/v1/deployments demo"}, def)
}

func TestMergeWorkloadsPreservesViews(t *testing.T) {
	existing := []byte(`views:
  v1/pods:
    columns:
      - NAME
      - STATUS
workloads:
  default:
    - v1/services
  other:
    - v1/pods
`)

	out, err := mergeWorkloads(existing, []string{"v1/pods", "batch/v1/jobs default"})
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(out, &doc))

	// views section preserved.
	views, ok := doc["views"].(map[string]any)
	require.True(t, ok, "views section must be preserved")
	_, ok = views["v1/pods"]
	assert.True(t, ok, "view entry must be preserved")

	wl, ok := doc["workloads"].(map[string]any)
	require.True(t, ok)
	// default replaced.
	assert.Equal(t, []any{"v1/pods", "batch/v1/jobs default"}, wl["default"])
	// other workload set preserved.
	assert.Equal(t, []any{"v1/pods"}, wl["other"])
}

func TestIsStandardGroup(t *testing.T) {
	assert.True(t, isStandardGroup("v1"))
	assert.True(t, isStandardGroup("apps/v1"))
	assert.True(t, isStandardGroup("batch/v1"))
	assert.True(t, isStandardGroup("networking.k8s.io/v1"))
	assert.False(t, isStandardGroup("examples.demo.io/v1"))
	assert.False(t, isStandardGroup("monitoring.coreos.com/v1"))
}
