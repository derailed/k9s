package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPort(t *testing.T) {
	uu := map[string]struct {
		port, e string
	}{
		"empty": {
			"", "",
		},
		"full": {
			"co/fred:8000", "8000",
		},
		"named": {
			"fred:8000", "8000",
		},
		"port": {
			"8000", "8000",
		},
		"protocol": {
			"dns:53╱UDP", "53",
		},
		"unamed": {
			"dns/53", "53",
		},
		"pod-dashed": {
			"blee-fred/:5000", "5000",
		},
		"co-dashed": {
			"blee/fred-doh:5000", "5000",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, extractPort(u.port))
		})
	}
}

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
