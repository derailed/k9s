package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractContainer(t *testing.T) {
	uu := map[string]struct {
		port, e string
	}{
		"full": {
			"co/port:8000", "co",
		},
		"unamed": {
			"co/:8000", "co",
		},
		"protocol": {
			"co/dns:53╱UDP", "co",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, extractContainer(u.port))
		})
	}
}
