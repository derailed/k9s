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
		{"", "", ""},
		{resource.MissingValue, "", delta()},
		{resource.NAValue, "", delta()},
		{"fred", "fred", ""},
		{"fred", "blee", delta()},
		{"1", "1", ""},
		{"1", "2", plus()},
		{"2", "1", minus()},
		{"2m33s", "2m33s", ""},
		{"2m33s", "1m", minus()},
		{"33s", "1m", plus()},
		{"10Gi", "10Gi", ""},
		{"10Gi", "20Gi", plus()},
		{"30Gi", "20Gi", minus()},
		{"15%", "15%", ""},
		{"20%", "40%", plus()},
		{"5%", "2%", minus()},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, deltas(u.s1, u.s2))
	}
}

func TestToPerc(t *testing.T) {
	assert.Equal(t, "50%", toPerc(50))
}
