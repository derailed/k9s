// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestHasAll(t *testing.T) {
	uu := map[string]struct {
		scopes []string
		e      bool
	}{
		"empty":  {},
		"all":    {scopes: []string{"blee", "duh", AllScopes}, e: true},
		"no-all": {scopes: []string{"blee", "duh", "alla"}},
	}

	for k := range uu {
		u := uu[k]
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
		"yes":   {s: "blee", ss: []string{"yo", "duh", "blee"}, e: true},
		"no":    {s: "blue", ss: []string{"yo", "duh", "blee"}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, includes(u.ss, u.s))
		})
	}
}

func TestInScope(t *testing.T) {
	uu := map[string]struct {
		ss []string
		aa map[string]struct{}
		e  bool
	}{
		"empty":         {},
		"yes":           {e: true, ss: []string{"blee", "duh", "fred"}, aa: map[string]struct{}{"blee": {}, "fred": {}, "duh": {}}},
		"no":            {ss: []string{"blee", "duh", "fred"}, aa: map[string]struct{}{"blee1": {}, "fred1": {}}},
		"empty scopes":  {aa: map[string]struct{}{"blee1": {}, "fred1": {}}},
		"empty aliases": {ss: []string{"blee1", "fred1"}},
		"all":           {e: true, ss: []string{AllScopes}, aa: map[string]struct{}{"blee1": {}, "fred1": {}}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, inScope(u.ss, u.aa))
		})
	}
}
