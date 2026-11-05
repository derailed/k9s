// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
)

func TestBrowserSuggestFilter(t *testing.T) {
	b := NewBrowser(client.PodGVR).(*Browser)
	b.app = NewApp(mock.NewMockConfig(t))
	b.app.filterHistory.Push("-l app=MyApp")
	b.app.filterHistory.Push("nginx")

	uu := map[string]struct {
		s string
		e sort.StringSlice
	}{
		"empty": {
			e: sort.StringSlice{"-l app=MyApp", "nginx"},
		},
		"label": {
			s: "-l app=My",
			e: sort.StringSlice{"App"},
		},
		"label-other-case": {
			s: "-l app=my",
		},
		"filter": {
			s: "ng",
			e: sort.StringSlice{"inx"},
		},
		"same": {
			s: "nginx",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, b.suggestFilter()(u.s))
		})
	}
}
