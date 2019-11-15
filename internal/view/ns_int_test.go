package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNSCleanser(t *testing.T) {
	var v Namespace

	uu := []struct {
		s, e string
	}{
		{"fred", "fred"},
		{"fred+", "fred"},
		{"fred(*)", "fred"},
		{"fred+(*)", "fred"},
		{"fred-blee+(*)", "fred-blee"},
		{"fred1-blee2+(*)", "fred1-blee2"},
		{"fred(𝜟)", "fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, v.cleanser(u.s))
	}
}
