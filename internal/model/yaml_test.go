package model

import (
	"testing"

	"github.com/sahilm/fuzzy"
	"github.com/stretchr/testify/assert"
)

func TestYAML_rxFilter(t *testing.T) {
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
					MatchedIndexes: []int{0, 3},
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
					MatchedIndexes: []int{0, 3},
				},
				{
					Str:            "foo",
					Index:          2,
					MatchedIndexes: []int{0, 3},
				},
				{
					Str:            "foo",
					Index:          2,
					MatchedIndexes: []int{8, 11},
				},
			},
		},
	}
	var y YAML
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, y.rxFilter(u.q, u.lines))
		})
	}
}
