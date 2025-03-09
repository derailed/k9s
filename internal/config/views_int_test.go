// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomView_getVS(t *testing.T) {
	uu := map[string]struct {
		cv      *CustomView
		gvr, ns string
		e       *ViewSetting
	}{
		"empty": {},

		"gvr": {
			gvr: "v1/pods",
			e: &ViewSetting{
				Columns: []string{"NAMESPACE", "NAME", "AGE", "IP"},
			},
		},

		"gvr+ns": {
			gvr: "v1/pods",
			ns:  "default",
			e: &ViewSetting{
				Columns: []string{"NAME", "IP", "AGE"},
			},
		},

		"rx": {
			gvr: "v1/pods",
			ns:  "ns-fred",
			e: &ViewSetting{
				Columns: []string{"AGE", "NAME", "IP"},
			},
		},

		"toast-no-ns": {
			gvr: "v1/pods",
			ns:  "zorg",
		},

		"toast-no-res": {
			gvr: "v1/services",
			ns:  "zorg",
		},
	}

	v := NewCustomView()
	assert.NoError(t, v.Load("testdata/views/views.yaml"))
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, v.getVS(u.gvr, u.ns))
		})
	}
}
