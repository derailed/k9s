// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomView_getVS(t *testing.T) {
	uu := map[string]struct {
		cv      *CustomView
		gvr, ns string
		e       *ViewSetting
	}{
		"empty": {},

		"miss": {
			gvr: "zorg",
		},

		"gvr": {
			gvr: client.PodGVR.String(),
			e: &ViewSetting{
				Columns: []string{"NAMESPACE", "NAME", "AGE", "IP"},
			},
		},

		"gvr+ns": {
			gvr: client.PodGVR.String(),
			ns:  "default",
			e: &ViewSetting{
				Columns: []string{"NAME", "IP", "AGE"},
			},
		},

		"rx": {
			gvr: client.PodGVR.String(),
			ns:  "ns-fred",
			e: &ViewSetting{
				Columns: []string{"AGE", "NAME", "IP"},
			},
		},

		"alias": {
			gvr: "bozo",
			e: &ViewSetting{
				Columns: []string{"DUH", "BLAH", "BLEE"},
			},
		},

		"toast-no-ns": {
			gvr: client.PodGVR.String(),
			ns:  "zorg",
			e: &ViewSetting{
				Columns: []string{"NAMESPACE", "NAME", "AGE", "IP"},
			},
		},

		"toast-no-res": {
			gvr: client.SvcGVR.String(),
			ns:  "zorg",
		},
	}

	v := NewCustomView()
	require.NoError(t, v.Load("testdata/views/views.yaml"))
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, v.getVS(u.gvr, u.ns))
		})
	}
}
