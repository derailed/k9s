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
		{"fred", "fred", "fred"},
		{"fred", "blee", delta("blee")},
		{"1", "2", plus("2")},
		{"2", "1", minus("1")},
		{"10Gi", "20Gi", plus("20Gi")},
		{"15%(-)", "15%", "15%"},
		{resource.MissingValue, "fred", delta("fred")},
		{resource.NAValue, "fred", delta("fred")},
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
