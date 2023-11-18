// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart_test

import (
	"strconv"
	"testing"

	"github.com/derailed/k9s/internal/tchart"
	"github.com/stretchr/testify/assert"
)

func TestDial3x3(t *testing.T) {
	d := tchart.NewDotMatrix()
	for n := 0; n <= 2; n++ {
		i := n
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			assert.Equal(t, tchart.To3x3Char(i), d.Print(i))
		})
	}
}
