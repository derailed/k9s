// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestScansShouldExclude(t *testing.T) {
	uu := map[string]struct {
		sc config.ImageScans
		ns string
		ll map[string]string
		e  bool
	}{
		"empty": {
			sc: config.NewImageScans(),
		},
		"exclude-ns": {
			sc: config.ImageScans{
				Enable: true,
				Exclusions: config.ScanExcludes{
					Namespaces: []string{"ns-1", "ns-2", "ns-3"},
					Labels: config.Labels{
						"app": []string{"fred", "blee"},
					},
				},
			},
			ns: "ns-1",
			ll: map[string]string{
				"app": "freddy",
			},
			e: true,
		},
		"include-ns": {
			sc: config.ImageScans{
				Enable: true,
				Exclusions: config.ScanExcludes{
					Namespaces: []string{"ns-1", "ns-2", "ns-3"},
					Labels: config.Labels{
						"app": []string{"fred", "blee"},
					},
				},
			},
			ns: "ns-4",
			ll: map[string]string{
				"app": "bozo",
			},
		},
		"exclude-labels": {
			sc: config.ImageScans{
				Enable: true,
				Exclusions: config.ScanExcludes{
					Namespaces: []string{"ns-1", "ns-2", "ns-3"},
					Labels: config.Labels{
						"app": []string{"fred", "blee"},
					},
				},
			},
			ns: "ns-4",
			ll: map[string]string{
				"app": "fred",
			},
			e: true,
		},
		"include-labels": {
			sc: config.ImageScans{
				Enable: true,
				Exclusions: config.ScanExcludes{
					Namespaces: []string{"ns-1", "ns-2", "ns-3"},
					Labels: config.Labels{
						"app": []string{"fred", "blee"},
					},
				},
			},
			ns: "ns-4",
			ll: map[string]string{
				"app": "freddy",
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.sc.ShouldExclude(u.ns, u.ll))
		})
	}
}
