// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/xray"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXrayGotoCmdBeforeTreeLoaded(t *testing.T) {
	v, ok := NewXray(client.NewGVR("deployments")).(*Xray)
	assert.True(t, ok)
	v.CmdBuff().SetActive(true)

	assert.NotPanics(t, func() {
		v.gotoCmd(nil)
	})
	assert.False(t, v.CmdBuff().IsActive())
}

func TestXrayK9sEnvFilter(t *testing.T) {
	uu := map[string]struct {
		spec          *xray.NodeSpec
		filter        string
		resetFilter   bool
		wantHasFilter bool
		wantFilter    string
		wantRawFilter string
		wantNS        string
		wantName      string
		wantPod       string
		wantContainer string
	}{
		"pod-no-filter-falls-back-to-selected-name": {
			spec: &xray.NodeSpec{
				GVRs:     client.GVRs{client.PodGVR},
				Paths:    []string{"kube-system/coredns-559f6c778d-z5d7w"},
				Statuses: []string{xray.OkStatus},
			},
			wantHasFilter: true,
			wantFilter:    "coredns-559f6c778d-z5d7w",
			wantRawFilter: "",
			wantNS:        "kube-system",
			wantName:      "coredns-559f6c778d-z5d7w",
		},
		"pod-active-filter": {
			spec: &xray.NodeSpec{
				GVRs:     client.GVRs{client.PodGVR},
				Paths:    []string{"kube-system/coredns-559f6c778d-z5d7w"},
				Statuses: []string{xray.OkStatus},
			},
			filter:        "coredns",
			wantHasFilter: true,
			wantFilter:    "coredns",
			wantRawFilter: "coredns",
			wantNS:        "kube-system",
			wantName:      "coredns-559f6c778d-z5d7w",
		},
		"container-no-filter": {
			spec: &xray.NodeSpec{
				GVRs:     client.GVRs{client.CoGVR, client.PodGVR},
				Paths:    []string{"kube-system/coredns", "kube-system/coredns-559f6c778d-z5d7w"},
				Statuses: []string{xray.OkStatus, xray.OkStatus},
			},
			wantHasFilter: true,
			wantFilter:    "coredns",
			wantRawFilter: "",
			wantNS:        "kube-system",
			wantName:      "coredns",
			wantPod:       "coredns-559f6c778d-z5d7w",
			wantContainer: "coredns",
		},
		"container-active-filter": {
			spec: &xray.NodeSpec{
				GVRs:     client.GVRs{client.CoGVR, client.PodGVR},
				Paths:    []string{"kube-system/coredns", "kube-system/coredns-559f6c778d-z5d7w"},
				Statuses: []string{xray.OkStatus, xray.OkStatus},
			},
			filter:        "-f core",
			wantHasFilter: true,
			wantFilter:    "-f core",
			wantRawFilter: "-f core",
			wantNS:        "kube-system",
			wantName:      "coredns",
			wantPod:       "coredns-559f6c778d-z5d7w",
			wantContainer: "coredns",
		},
		"reset-filter-restores-empty-raw-filter": {
			spec: &xray.NodeSpec{
				GVRs:     client.GVRs{client.PodGVR},
				Paths:    []string{"kube-system/coredns-559f6c778d-z5d7w"},
				Statuses: []string{xray.OkStatus},
			},
			filter:        "coredns",
			resetFilter:   true,
			wantHasFilter: true,
			wantFilter:    "coredns-559f6c778d-z5d7w",
			wantRawFilter: "",
			wantNS:        "kube-system",
			wantName:      "coredns-559f6c778d-z5d7w",
		},
		"no-selection": {
			filter: "coredns",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			v, ok := NewXray(client.PodGVR).(*Xray)
			require.True(t, ok)
			v.app = makeEnvApp(t)
			v.envFn = v.k9sEnv

			if u.spec != nil {
				node := tview.NewTreeNode(u.spec.Path()).SetReference(*u.spec)
				v.SetRoot(node)
				v.SetCurrentNode(node)
			}
			if u.filter != "" {
				v.CmdBuff().SetText(u.filter, "", true)
			}
			if u.resetFilter {
				v.CmdBuff().Reset()
			}

			env := v.EnvFn()()
			if !u.wantHasFilter {
				assert.NotContains(t, env, "FILTER")
				assert.NotContains(t, env, "RAW_FILTER")
				return
			}

			assert.Contains(t, env, "RAW_FILTER")
			assert.Equal(t, u.wantRawFilter, env["RAW_FILTER"])
			assert.Equal(t, u.wantFilter, env["FILTER"])
			assert.Equal(t, u.wantNS, env["NAMESPACE"])
			assert.Equal(t, u.wantName, env["NAME"])
			assert.Equal(t, u.wantPod, env["POD"])
			assert.Equal(t, u.wantContainer, env["CONTAINER"])

			sub, err := env.Substitute(`FILTER=[$FILTER] RAW_FILTER=[$RAW_FILTER]`)
			require.NoError(t, err)
			assert.Equal(t, "FILTER=["+u.wantFilter+"] RAW_FILTER=["+u.wantRawFilter+"]", sub)
		})
	}
}
