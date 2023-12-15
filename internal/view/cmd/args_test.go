// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagsNew(t *testing.T) {
	uu := map[string]struct {
		aa []string
		ll args
	}{
		"empty": {
			ll: make(args),
		},
		"ns": {
			aa: []string{"ns1"},
			ll: args{nsKey: "ns1"},
		},
		"ns-spaces": {
			aa: []string{" ns1 "},
			ll: args{nsKey: "ns1"},
		},
		"filter": {
			aa: []string{"/fred"},
			ll: args{filterKey: "fred"},
		},
		"inverse-filter": {
			aa: []string{"/!fred"},
			ll: args{filterKey: "!fred"},
		},
		"fuzzy-filter": {
			aa: []string{"-f", "fred"},
			ll: args{filterKey: "fred"},
		},
		"filter-ns": {
			aa: []string{"/fred", "  ns1 "},
			ll: args{nsKey: "ns1", filterKey: "fred"},
		},
		"label": {
			aa: []string{"app=fred"},
			ll: args{labelKey: "app=fred"},
		},
		"label-toast": {
			aa: []string{"="},
			ll: make(args),
		},
		"multi-labels": {
			aa: []string{"app=fred,blee=duh"},
			ll: args{labelKey: "app=fred,blee=duh"},
		},
		"label-ns": {
			aa: []string{"a=b,c=d", "  ns1  "},
			ll: args{labelKey: "a=b,c=d", nsKey: "ns1"},
		},
		"full-monty": {
			aa: []string{"app=fred", "ns1", "-f", "blee", "/zorg"},
			ll: args{filterKey: "zorg", labelKey: "app=fred", nsKey: "ns1"},
		},
		"full-monty=ctx": {
			aa: []string{"app=fred", "ns1", "-f", "blee", "/zorg", "@ctx1"},
			ll: args{filterKey: "zorg", labelKey: "app=fred", nsKey: "ns1", contextKey: "ctx1"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			l := newArgs(u.aa)
			assert.Equal(t, len(u.ll), len(l))
			assert.Equal(t, u.ll, l)
		})
	}
}

func TestFlagsHasFilters(t *testing.T) {
	uu := map[string]struct {
		aa []string
		ok bool
	}{
		"empty": {},
		"ns": {
			aa: []string{"ns1"},
		},
		"filter": {
			aa: []string{"/fred"},
			ok: true,
		},
		"inverse-filter": {
			aa: []string{"/!fred"},
			ok: true,
		},
		"fuzzy-filter": {
			aa: []string{"-f", "fred"},
			ok: true,
		},
		"filter-ns": {
			aa: []string{"/fred", "ns1"},
			ok: true,
		},
		"label": {
			aa: []string{"app=fred"},
			ok: true,
		},
		"multi-labels": {
			aa: []string{"app=fred,blee=duh"},
			ok: true,
		},
		"label-ns": {
			aa: []string{"app=fred", "ns1"},
			ok: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			l := newArgs(u.aa)
			assert.Equal(t, u.ok, l.hasFilters())
		})
	}
}
