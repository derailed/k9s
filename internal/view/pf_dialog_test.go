// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPort(t *testing.T) {
	uu := map[string]struct {
		portSpec, e string
	}{
		"full": {
			portSpec: "co::8000",
			e:        "8000",
		},
		"toast": {
			portSpec: "co:8000",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, extractPort(u.portSpec))
		})
	}
}
