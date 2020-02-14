package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPort(t *testing.T) {
	uu := map[string]struct {
		port, e string
	}{
		"full": {
			"fred:8000", "8000",
		},
		"port": {
			"8000", "8000",
		},
		"protocol": {
			"dns:53╱UDP", "53",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, extractPort(u.port))
		})
	}
}
