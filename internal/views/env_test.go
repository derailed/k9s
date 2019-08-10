package views

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
		"match":   {q: "$A", err: nil, e: "10"},
		"noMatch": {q: "$BLEE", err: errors.New("No matching for $BLEE"), e: ""},
		"lower":   {q: "$b", err: nil, e: "blee"},
		"dash":    {q: "$col-0", err: nil, e: "fred"},
	}

	e := K9sEnv{
		"A":     "10",
		"B":     "blee",
		"COL-0": "fred",
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			a, err := e.envFor(u.q)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, a)
		})
	}
}
