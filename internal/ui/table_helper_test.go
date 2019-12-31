package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLabelSelector(t *testing.T) {
	uu := map[string]struct {
		sel string
		e   bool
	}{
		"cool":       {"-l app=fred,env=blee", true},
		"noMode":     {"app=fred,env=blee", false},
		"noSpace":    {"-lapp=fred,env=blee", true},
		"wrongLabel": {"-f app=fred,env=blee", false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, IsLabelSelector(u.sel))
		})
	}
}

func TestTrimLabelSelector(t *testing.T) {
	uu := map[string]struct {
		sel, e string
	}{
		"cool":    {"-l app=fred,env=blee", "app=fred,env=blee"},
		"noSpace": {"-lapp=fred,env=blee", "app=fred,env=blee"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, TrimLabelSelector(u.sel))
		})
	}
}
