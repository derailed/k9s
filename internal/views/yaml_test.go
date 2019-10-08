package views

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
			`api: fred
		version: v1`,
			`[steelblue::b]api[white::-]: [papayawhip::]fred
		[steelblue::b]version[white::-]: [papayawhip::]v1`,
		},
		{
			`api:
		version: v1`,
			`[steelblue::b]api[white::-]:
		[steelblue::b]version[white::-]: [papayawhip::]v1`,
		},
		{
			"      fred:blee",
			"[papayawhip::]      fred:blee",
		},
		{
			"fred blee: blee",
			"[steelblue::b]fred blee[white::-]: [papayawhip::]blee",
		},
		{
			"Node-Selectors:  <none>",
			"[steelblue::b]Node-Selectors[white::-]: [papayawhip::] <none>",
		},
		{
			"fred.blee:  <none>",
			"[steelblue::b]fred.blee[white::-]: [papayawhip::] <none>",
		},
		{
			"certmanager.k8s.io/cluster-issuer: nameOfClusterIssuer",
			"[steelblue::b]certmanager.k8s.io/cluster-issuer[white::-]: [papayawhip::]nameOfClusterIssuer",
		},
		{
			"Message: Pod The node was low on resource: [DiskPressure].",
			"[steelblue::b]Message[white::-]: [papayawhip::]Pod The node was low on resource: [DiskPressure[].",
		},
	}

	s, _ := config.NewStyles("skins/stock.yml")
	for _, u := range uu {
		assert.Equal(t, u.e, colorizeYAML(s.Views().Yaml, u.s))
	}
}
