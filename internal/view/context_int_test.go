package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleaner(t *testing.T) {
	uu := map[string]struct {
		s, e string
	}{
		"normal":  {"fred", "fred"},
		"default": {"fred*", "fred"},
		"delta":   {"fred(𝜟)", "fred"},
	}

	v := Context{}
	for k := range uu {
   u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, v.cleanser(u.s))
		})
	}
}
