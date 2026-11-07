// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_contextMerge(t *testing.T) {
	uu := map[string]struct {
		c1, c2, e *Context
	}{
		"empty": {},
		"nil": {
			c1: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1", "ns2", "ns3"},
				},
			},
			e: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1", "ns2", "ns3"},
				},
			},
		},
		"deltas": {
			c1: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1", "ns2", "ns3"},
				},
			},
			c2: &Context{
				Namespace: &Namespace{
					Active:    "ns10",
					Favorites: []string{"ns10", "ns11", "ns12"},
				},
			},
			e: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1", "ns2", "ns3", "ns10", "ns11", "ns12"},
				},
			},
		},
		"recent": {
			c1: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1"},
					Recent:    []string{"ns2"},
				},
			},
			c2: &Context{
				Namespace: &Namespace{
					Active:    "ns10",
					Favorites: []string{"ns10"},
					Recent:    []string{"ns11"},
				},
			},
			e: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1", "ns10"},
					Recent:    []string{"ns2", "ns11"},
				},
			},
		},
		"no-namespace": {
			c1: NewContext(),
			c2: &Context{},
			e:  NewContext(),
		},
		"too-many-favs": {
			c1: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1", "ns2", "ns3", "ns4", "ns5", "ns6", "ns7", "ns8", "ns9"},
				},
			},
			c2: &Context{
				Namespace: &Namespace{
					Active:    "ns10",
					Favorites: []string{"ns10", "ns11", "ns12"},
				},
			},
			e: &Context{
				Namespace: &Namespace{
					Active:    "ns1",
					Favorites: []string{"ns1", "ns2", "ns3", "ns4", "ns5", "ns6", "ns7", "ns8", "ns9"},
				},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			u.c1.merge(u.c2)
			assert.Equal(t, u.e, u.c1)
		})
	}
}

func Test_nsMigrate(t *testing.T) {
	lockOn, lockOff := true, false
	uu := map[string]struct {
		ns, e *Namespace
	}{
		"none": {
			ns: &Namespace{Favorites: []string{"ns1"}},
			e:  &Namespace{Favorites: []string{"ns1"}},
		},
		"locked": {
			ns: &Namespace{LockFavorites: &lockOn, Favorites: []string{"ns1"}},
			e:  &Namespace{Favorites: []string{"ns1"}},
		},
		"unlocked": {
			ns: &Namespace{LockFavorites: &lockOff, Favorites: []string{"ns1"}},
			e:  &Namespace{Recent: []string{"ns1"}},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			u.ns.migrate()
			assert.Equal(t, u.e, u.ns)
		})
	}
}
