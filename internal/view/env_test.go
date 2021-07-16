package view

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvReplace(t *testing.T) {
	uu := map[string]struct {
		arg string
		err error
		e   string
	}{
		"no-args":   {arg: "blee blah", e: "blee blah"},
		"simple":    {arg: "$A", e: "10"},
		"substring": {arg: "$A and $AA", e: "10 and 20"},
		"with-text": {arg: "Something $A", e: "Something 10"},
		"noMatch":   {arg: "blah blah and $BLEE", err: errors.New(`no environment matching key "$BLEE":"BLEE"`), e: ""},
		"lower":     {arg: "And then $b happened", e: "And then blee happened"},
		"dash":      {arg: "$col0", e: "fred"},
		"mix":       {arg: "$col0 and then $a but $B", e: "fred and then 10 but blee"},
		"subs":      {arg: `{"spec" : {"suspend" : $COL0 }}`, e: `{"spec" : {"suspend" : fred }}`},
		"boolean":   {arg: "$COL-BOOL", e: "false"},
		"invert":    {arg: "$!COL-BOOL", e: "true"},
	}

	e := Env{
		"A":        "10",
		"AA":       "20",
		"B":        "blee",
		"COL0":     "fred",
		"FRED":     "fred",
		"COL-NAME": "zorg",
		"COL-BOOL": "false",
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			a, err := e.Substitute(u.arg)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, a)
		})
	}
}
