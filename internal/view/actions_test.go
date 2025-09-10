// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestHasAll(t *testing.T) {
	uu := map[string]struct {
		scopes []string
		e      bool
	}{
		"empty": {},

		"all": {
			scopes: []string{"blee", "duh", AllScopes},
			e:      true,
		},

		"none": {
			scopes: []string{"blee", "duh", "alla"},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, hasAll(u.scopes))
		})
	}
}

func TestIncludes(t *testing.T) {
	uu := map[string]struct {
		s  string
		ss []string
		e  bool
	}{
		"empty": {},

		"yes": {
			s:  "blee",
			ss: []string{"yo", "duh", "blee"},
			e:  true,
		},

		"no": {
			s:  "blue",
			ss: []string{"yo", "duh", "blee"},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, includes(u.ss, u.s))
		})
	}
}

func TestInScope(t *testing.T) {
	uu := map[string]struct {
		ss []string
		aa sets.Set[string]
		e  bool
	}{
		"empty": {},

		"yes": {
			e:  true,
			ss: []string{"blee", "duh", "fred"},
			aa: sets.New("blee", "fred", "duh"),
		},

		"no": {
			ss: []string{"blee", "duh", "fred"},
			aa: sets.New("blee1", "fred1"),
		},

		"no-scopes": {
			aa: sets.New("aa", "blee1", "fred1"),
		},

		"no-aliases": {
			ss: []string{"blee1", "fred1"},
		},

		"all": {
			e:  true,
			ss: []string{AllScopes},
			aa: sets.New("blee1", "fred1"),
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, inScope(u.ss, u.aa))
		})
	}
}
