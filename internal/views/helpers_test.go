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
		{"2m33s", "1m", minus()},
		{"10Gi", "20Gi", plus()},
		{"15%(-)", "15%", delta()},
		{resource.MissingValue, "", delta()},
		{resource.NAValue, "", delta()},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, deltas(u.s1, u.s2))
	}
}
