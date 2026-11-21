// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestYaml(t *testing.T) {
	uu := []struct {
		s, e string
	}{
		{
			"Events (2026-10-06 local):",
			"[#4682b4::b]Events (2026-10-06 local)[#ffffff::-]:",
		},
		{
			`api: fred
		   version: v1`,
			`[#4682b4::b]api[#ffffff::-]: [#ffefd5::]fred
		   [#4682b4::b]version[#ffffff::-]: [#ffefd5::]v1`,
		},
		{
			`api: <<<"search_0">>>fred<<<"">>>
		   version: v1`,
			`[#4682b4::b]api[#ffffff::-]: [#ffefd5::]["search_0"]fred[""]
		   [#4682b4::b]version[#ffffff::-]: [#ffefd5::]v1`,
		},
		{
			`api:
			version: v1`,
			`[#4682b4::b]api[#ffffff::-]:
			[#4682b4::b]version[#ffffff::-]: [#ffefd5::]v1`,
		},
		{
			"      fred:blee",
			"[#ffefd5::]      fred:blee",
		},
		{
			"fred blee: blee",
			"[#4682b4::b]fred blee[#ffffff::-]: [#ffefd5::]blee",
		},
		{
			"Node-Selectors:  <none>",
			"[#4682b4::b]Node-Selectors[#ffffff::-]: [#ffefd5::] <none>",
		},
		{
			"fred.blee:  <none>",
			"[#4682b4::b]fred.blee[#ffffff::-]: [#ffefd5::] <none>",
		},
		{
			"certmanager.k8s.io/cluster-issuer: nameOfClusterIssuer",
			"[#4682b4::b]certmanager.k8s.io/cluster-issuer[#ffffff::-]: [#ffefd5::]nameOfClusterIssuer",
		},
		{
			"Message: Pod The node was low on resource: [DiskPressure].",
			"[#4682b4::b]Message[#ffffff::-]: [#ffefd5::]Pod The node was low on resource: [DiskPressure[].",
		},
		{
			`data: "<<<"`,
			`[#4682b4::b]data[#ffffff::-]: [#ffefd5::]"<<<"`,
		},
	}

	s := config.NewStyles()
	for _, u := range uu {
		assert.Equal(t, u.e, colorizeYAML(s.Views().Yaml, u.s))
	}
}

func TestWarningRowColor(t *testing.T) {
	style := config.NewStyles().Views().Yaml
	raw := "Events (2026-10-08 local):\n  ⚠ 07:45:44 → ::47 (x24)  Unhealthy  kubelet  probe failed [red]\n  ⚠     connection refused\n    07:45:48  Started  kubelet  container started"
	got := colorizeYAML(style, raw)
	assert.Contains(t, got, "[#4682b4::b]Events (2026-10-08 local)[#ffffff::-]:")
	assert.Contains(t, got, "[orange::]  ⚠ 07:45:44 → ::47 (x24)  Unhealthy  kubelet  probe failed [red[]")
	assert.Contains(t, got, "[orange::]  ⚠     connection refused")
	assert.Contains(t, got, "[#ffefd5::]    07:45:48  Started")
}
