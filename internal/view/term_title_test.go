// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTermTitle(t *testing.T) {
	uu := map[string]struct {
		ctx, view, e string
	}{
		"full": {
			ctx:  "prod",
			view: "Pods",
			e:    "k9s | prod | pods",
		},
		"spaces": {
			ctx:  "prod",
			view: "Port Forward",
			e:    "k9s | prod | portforward",
		},
		"no-view": {
			ctx: "prod",
			e:   "k9s | prod",
		},
		"no-context": {
			view: "contexts",
			e:    "k9s | contexts",
		},
		"empty": {
			e: "k9s",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, formatTermTitle(u.ctx, u.view))
		})
	}
}
