// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestDeltas(t *testing.T) {
	uu := []struct {
		s1, s2, e string
	}{
		{"", "", ""},
		{render.MissingValue, "", DeltaSign},
		{render.NAValue, "", ""},
		{"fred", "fred", ""},
		{"fred", "blee", DeltaSign},
		{"1", "1", ""},
		{"1", "2", PlusSign},
		{"2", "1", MinusSign},
		{"2m33s", "2m33s", ""},
		{"2m33s", "1m", MinusSign},
		{"33s", "1m", PlusSign},
		{"10Gi", "10Gi", ""},
		{"10Gi", "20Gi", PlusSign},
		{"30Gi", "20Gi", MinusSign},
		{"15%", "15%", ""},
		{"20%", "40%", PlusSign},
		{"5%", "2%", MinusSign},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, Deltas(u.s1, u.s2))
	}
}
