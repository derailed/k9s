// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"strconv"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/sahilm/fuzzy"
	"github.com/stretchr/testify/assert"
)

func matchTag(i int, s string) string {
	return `<<<"search_` + strconv.Itoa(i) + `">>>` + s + `<<<"">>>`
}

func TestLiveViewSetText(t *testing.T) {
	s := `
apiVersion: v1
  data:
    the secret name you want to quote to use tls.","title":"secretName","type":"string"}},"required":["http","class","classInSpec"],"type":"object"}
`

	v := NewLiveView(NewApp(config.NewConfig(nil)), "fred", nil)
	assert.NoError(t, v.Init(context.Background()))
	v.text.SetText(colorizeYAML(config.Yaml{}, s))

	assert.Equal(t, s, sanitizeEsc(v.text.GetText(true)))
}

func TestLiveView_linesWithRegions(t *testing.T) {
	uu := map[string]struct {
		lines   []string
		matches fuzzy.Matches
		e       []string
	}{
		"empty-lines": {
			e: []string{},
		},
		"no-match": {
			lines: []string{"bar"},
			e:     []string{"bar"},
		},
		"single-match": {
			lines: []string{"foo", "bar", "baz"},
			matches: fuzzy.Matches{
				{Index: 1, MatchedIndexes: []int{0, 3}},
			},
			e: []string{"foo", matchTag(0, "bar"), "baz"},
		},
		"multiple-matches": {
			lines: []string{"foo", "bar", "baz"},
			matches: fuzzy.Matches{
				{Index: 1, MatchedIndexes: []int{0, 3}},
				{Index: 2, MatchedIndexes: []int{0, 3}},
			},
			e: []string{"foo", matchTag(0, "bar"), matchTag(1, "baz")},
		},
		"multiple-matches-same-line": {
			lines: []string{"foosfoo baz", "dfbarfoos bar"},
			matches: fuzzy.Matches{
				{Index: 0, MatchedIndexes: []int{0, 3}},
				{Index: 0, MatchedIndexes: []int{4, 7}},
				{Index: 1, MatchedIndexes: []int{5, 8}},
			},
			e: []string{
				matchTag(0, "foo") + "s" + matchTag(1, "foo") + " baz",
				"dfbar" + matchTag(2, "foo") + "s bar",
			},
		},
	}
	var v LiveView
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, u.e, v.linesWithRegions(u.lines, u.matches))
		})
	}
}
