package views

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestDeltas(t *testing.T) {
	uu := []struct {
		s1, s2, e string
	}{
		{"", "", ""},
		{resource.MissingValue, "", deltaSign},
		{resource.NAValue, "", ""},
		{"fred", "fred", ""},
		{"fred", "blee", deltaSign},
		{"1", "1", ""},
		{"1", "2", plusSign},
		{"2", "1", minusSign},
		{"2m33s", "2m33s", ""},
		{"2m33s", "1m", minusSign},
		{"33s", "1m", plusSign},
		{"10Gi", "10Gi", ""},
		{"10Gi", "20Gi", plusSign},
		{"30Gi", "20Gi", minusSign},
		{"15%", "15%", ""},
		{"20%", "40%", plusSign},
		{"5%", "2%", minusSign},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, deltas(u.s1, u.s2))
	}
}
