// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestLogOptionsToggleAllContainers(t *testing.T) {
	uu := map[string]struct {
		opts dao.LogOptions
		co   string
		want bool
	}{
		"empty": {
			opts: dao.LogOptions{},
			want: true,
		},
		"container": {
			opts: dao.LogOptions{Container: "blee"},
			want: true,
		},
		"default-container": {
			opts: dao.LogOptions{AllContainers: true},
			co:   "blee",
		},
		"single-container": {
			opts: dao.LogOptions{Container: "blee", SingleContainer: true},
			co:   "blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.opts.DefaultContainer = "blee"
			u.opts.ToggleAllContainers()
			assert.Equal(t, u.want, u.opts.AllContainers)
			assert.Equal(t, u.co, u.opts.Container)
		})
	}
}
