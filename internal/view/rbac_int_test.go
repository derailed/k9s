package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasVerb(t *testing.T) {
	uu := []struct {
		vv []string
		v  string
		e  bool
	}{
		{[]string{"*"}, "get", true},
		{[]string{"get", "list", "watch"}, "watch", true},
		{[]string{"get", "dope", "list"}, "watch", false},
		{[]string{"get"}, "get", true},
		{[]string{"post"}, "create", true},
		{[]string{"put"}, "update", true},
		{[]string{"list", "deletecollection"}, "deletecollection", true},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, hasVerb(u.vv, u.v))
	}
}

func TestAsVerbs(t *testing.T) {
	ok, nok := toVerbIcon(true), toVerbIcon(false)

	uu := []struct {
		vv, e []string
	}{
		{
			[]string{"*"},
			[]string{ok, ok, ok, ok, ok, ok, ok, ok, ""},
		},
		{
			[]string{"get", "list", "patch"},
			[]string{ok, ok, nok, nok, ok, nok, nok, nok, ""},
		},
		{
			[]string{"get", "list", "deletecollection", "post"},
			[]string{ok, ok, nok, ok, nok, nok, nok, ok, ""},
		},
		{
			[]string{"get", "list", "blee"},
			[]string{ok, ok, nok, nok, nok, nok, nok, nok, "blee"},
		},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, asVerbs(u.vv))
	}
}
