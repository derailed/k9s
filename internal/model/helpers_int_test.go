// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"testing"

	"github.com/sahilm/fuzzy"
	"github.com/stretchr/testify/assert"
)

func Test_rxFilter(t *testing.T) {
	uu := map[string]struct {
		q     string
		lines []string
		e     fuzzy.Matches
	}{
		"empty-lines": {
			q: "foo",
			e: fuzzy.Matches{},
		},
		"no-match": {
			q:     "foo",
			lines: []string{"bar"},
			e:     fuzzy.Matches{},
		},
		"single-match": {
			q:     "foo",
			lines: []string{"foo", "bar", "baz"},
			e: fuzzy.Matches{
				{
					Str:            "foo",
					Index:          0,
					MatchedIndexes: []int{0, 1, 2},
				},
			},
		},
		"start-rx-match": {
			q:     "(?i)^foo",
			lines: []string{"foo", "fob", "barfoo"},
			e: fuzzy.Matches{
				{
					Str:            "(?i)^foo",
					Index:          0,
					MatchedIndexes: []int{0, 1, 2},
				},
			},
		},
		"end-rx-match": {
			q:     "foo$",
			lines: []string{"foo", "fob", "barfoo"},
			e: fuzzy.Matches{
				{
					Str:            "foo$",
					Index:          0,
					MatchedIndexes: []int{0, 1, 2},
				},
				{
					Str:            "foo$",
					Index:          2,
					MatchedIndexes: []int{3, 4, 5},
				},
			},
		},
		"multiple-matches": {
			q:     "foo",
			lines: []string{"foo", "bar", "foo bar foo", "baz"},
			e: fuzzy.Matches{
				{
					Str:            "foo",
					Index:          0,
					MatchedIndexes: []int{0, 1, 2},
				},
				{
					Str:            "foo",
					Index:          2,
					MatchedIndexes: []int{0, 1, 2},
				},
				{
					Str:            "foo",
					Index:          2,
					MatchedIndexes: []int{8, 9, 10},
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, rxFilter(u.q, u.lines))
		})
	}
}
