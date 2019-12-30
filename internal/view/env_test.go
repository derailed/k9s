package view

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestK9sEnv(t *testing.T) {

	uu := map[string]struct {
		q   string
		err error
		e   string
	}{
		"match":   {q: "$A", e: "10"},
		"noMatch": {q: "$BLEE", err: errors.New("No matching for $BLEE"), e: ""},
		"lower":   {q: "$b", e: "blee"},
		"dash":    {q: "$col0", e: "fred"},
		"mix":     {q: "$col0-blee", e: "fred-blee"},
	}

	e := K9sEnv{
		"A":    "10",
		"B":    "blee",
		"COL0": "fred",
	}

	for k := range uu {
   u := uu[k]
		t.Run(k, func(t *testing.T) {
			a, err := e.envFor(u.q)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, a)
		})
	}
}
