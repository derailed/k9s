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
		"deltas-locked": {
			c1: &Context{
				Namespace: &Namespace{
					Active:        "ns1",
					LockFavorites: true,
					Favorites:     []string{"ns1", "ns2", "ns3"},
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
					Active:        "ns1",
					LockFavorites: true,
					Favorites:     []string{"ns1", "ns2", "ns3"},
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
