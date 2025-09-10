package tchart_test

import (
	"strconv"
	"testing"

	"github.com/derailed/k9s/internal/tchart"
	"github.com/stretchr/testify/assert"
)

func TestDial3x3(t *testing.T) {
	d := tchart.NewDotMatrix()
	for n := range 2 {
		i := n
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			assert.Equal(t, tchart.To3x3Char(i), d.Print(i))
		})
	}
}
