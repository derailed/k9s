// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestLogIndicatorRefresh(t *testing.T) {
	defaults := config.NewStyles()
	uu := map[string]struct {
		li *view.LogIndicator
		e  string
	}{
		"all-containers": {
			view.NewLogIndicator(config.NewConfig(nil), defaults, true), "[::b]AllContainers:[gray::d]Off[-::]     [::b]Autoscroll:[limegreen::b]On[-::]      [::b]FullScreen:[gray::d]Off[-::]     [::b]Timestamps:[gray::d]Off[-::]     [::b]Wrap:[gray::d]Off[-::]\n",
		},
		"plain": {
			view.NewLogIndicator(config.NewConfig(nil), defaults, false), "[::b]Autoscroll:[limegreen::b]On[-::]      [::b]FullScreen:[gray::d]Off[-::]     [::b]Timestamps:[gray::d]Off[-::]     [::b]Wrap:[gray::d]Off[-::]\n",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.li.Refresh()
			assert.Equal(t, u.e, u.li.GetText(false))
		})
	}
}

func BenchmarkLogIndicatorRefresh(b *testing.B) {
	defaults := config.NewStyles()
	v := view.NewLogIndicator(config.NewConfig(nil), defaults, true)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v.Refresh()
	}
}
