package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPerc(t *testing.T) {
	uu := []struct {
		v1, v2, e float64
	}{
		{0, 0, 0},
		{100, 200, 50},
		{200, 100, 200},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, toPerc(u.v1, u.v2))
	}
}
