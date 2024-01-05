// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagsNew(t *testing.T) {
	uu := map[string]struct {
		i  *Interpreter
		aa []string
		ll args
	}{
		"empty": {
			i:  NewInterpreter("po"),
			ll: make(args),
		},
		"ns": {
			i:  NewInterpreter("po"),
			aa: []string{"ns1"},
			ll: args{nsKey: "ns1"},
		},
		"ns+spaces": {
			i:  NewInterpreter("po"),
			aa: []string{" ns1 "},
			ll: args{nsKey: "ns1"},
		},
		"filter": {
			i:  NewInterpreter("po"),
			aa: []string{"/fred"},
			ll: args{filterKey: "fred"},
		},
		"inverse-filter": {
			i:  NewInterpreter("po"),
			aa: []string{"/!fred"},
			ll: args{filterKey: "!fred"},
		},
		"fuzzy-filter": {
			i:  NewInterpreter("po"),
			aa: []string{"-f", "fred"},
			ll: args{fuzzyKey: "fred"},
		},
		"fuzzy-filter-nospace": {
			i:  NewInterpreter("po"),
			aa: []string{"-ffred"},
			ll: args{fuzzyKey: "fred"},
		},
		"filter+ns": {
			i:  NewInterpreter("po"),
			aa: []string{"/fred", "  ns1 "},
			ll: args{nsKey: "ns1", filterKey: "fred"},
		},
		"label": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred"},
			ll: args{labelKey: "app=fred"},
		},
		"label-toast": {
			i:  NewInterpreter("po"),
			aa: []string{"="},
			ll: make(args),
		},
		"multi-labels": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred,blee=duh"},
			ll: args{labelKey: "app=fred,blee=duh"},
		},
		"label+ns": {
			i:  NewInterpreter("po"),
			aa: []string{"a=b,c=d", "  ns1  "},
			ll: args{labelKey: "a=b,c=d", nsKey: "ns1"},
		},
		"full-monty": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred", "ns1", "-f", "blee", "/zorg"},
			ll: args{
				filterKey: "zorg",
				fuzzyKey:  "blee",
				labelKey:  "app=fred",
				nsKey:     "ns1",
			},
		},
		"full-monty+ctx": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred", "ns1", "-f", "blee", "/zorg", "@ctx1"},
			ll: args{
				filterKey:  "zorg",
				fuzzyKey:   "blee",
				labelKey:   "app=fred",
				nsKey:      "ns1",
				contextKey: "ctx1",
			},
		},
		"caps": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred", "ns1", "-f", "blee", "/zorg", "@Dev"},
			ll: args{
				filterKey:  "zorg",
				fuzzyKey:   "blee",
				labelKey:   "app=fred",
				nsKey:      "ns1",
				contextKey: "Dev"},
		},
		"ctx": {
			i:  NewInterpreter("ctx"),
			aa: []string{"Dev"},
			ll: args{contextKey: "Dev"},
		},
		"bork": {
			i:  NewInterpreter("apply -f"),
			ll: args{},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			l := newArgs(u.i, u.aa)
			assert.Equal(t, len(u.ll), len(l))
			assert.Equal(t, u.ll, l)
		})
	}
}

func TestFlagsHasFilters(t *testing.T) {
	uu := map[string]struct {
		i  *Interpreter
		aa []string
		ok bool
	}{
		"empty": {},
		"ns": {
			i:  NewInterpreter("po"),
			aa: []string{"ns1"},
		},
		"filter": {
			i:  NewInterpreter("po"),
			aa: []string{"/fred"},
			ok: true,
		},
		"inverse-filter": {
			i:  NewInterpreter("po"),
			aa: []string{"/!fred"},
			ok: true,
		},
		"fuzzy-filter": {
			i:  NewInterpreter("po"),
			aa: []string{"-f", "fred"},
			ok: true,
		},
		"filter+ns": {
			i:  NewInterpreter("po"),
			aa: []string{"/fred", "ns1"},
			ok: true,
		},
		"label": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred"},
			ok: true,
		},
		"multi-labels": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred,blee=duh"},
			ok: true,
		},
		"label+ns": {
			i:  NewInterpreter("po"),
			aa: []string{"app=fred", "ns1"},
			ok: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			l := newArgs(u.i, u.aa)
			assert.Equal(t, u.ok, l.hasFilters())
		})
	}
}
