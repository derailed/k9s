package views

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestDeltas(t *testing.T) {
	uu := []struct {
		s1, s2, e string
	}{
		{"fred", "fred", ""},
		{"fred", "blee", delta()},
		{"1", "2", plus()},
		{"2", "1", minus()},
		{"10Gi", "20Gi", plus()},
		{"15%(-)", "15%", ""},
		{resource.MissingValue, "", delta()},
		{resource.NAValue, "", delta()},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, deltas(u.s1, u.s2))
	}
}

func TestIsAlpha(t *testing.T) {
	uu := []struct {
		i string
		e bool
	}{
		{"fred", false},
		{"1Gi", true},
		{"1", true},
		{"", false},
		{resource.MissingValue, false},
		{resource.NAValue, false},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, isAlpha(u.i))
	}
}
