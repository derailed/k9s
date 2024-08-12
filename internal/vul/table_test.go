// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_sort(t *testing.T) {
	uu := map[string]struct {
		t1, t2 *table
	}{
		"simple": {
			t1: makeTable(t, "testdata/sort/no_dups/sc1.text"),
			t2: makeTable(t, "testdata/sort/no_dups/sc2.text"),
		},
		"dups": {
			t1: makeTable(t, "testdata/sort/dups/sc1.text"),
			t2: makeTable(t, "testdata/sort/dups/sc2.text"),
		},
		"full": {
			t1: makeTable(t, "testdata/sort/full/sc1.text"),
			t2: makeTable(t, "testdata/sort/full/sc2.text"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.t1.sort()
			assert.Equal(t, u.t2, u.t1)
		})
	}
}

func Test_sortSev(t *testing.T) {
	uu := map[string]struct {
		t1, t2 *table
	}{
		"simple": {
			t1: makeTable(t, "testdata/sort_sev/no_dups/sc1.text"),
			t2: makeTable(t, "testdata/sort_sev/no_dups/sc2.text"),
		},
		"dups": {
			t1: makeTable(t, "testdata/sort_sev/dups/sc1.text"),
			t2: makeTable(t, "testdata/sort_sev/dups/sc2.text"),
		},
		"full": {
			t1: makeTable(t, "testdata/sort_sev/full/sc1.text"),
			t2: makeTable(t, "testdata/sort_sev/full/sc2.text"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.t1.sortSev()
			assert.Equal(t, u.t2, u.t1)
		})
	}
}

// Helpers...

func makeTable(t *testing.T, path string) *table {
	f, err := os.Open(path)
	defer func() {
		_ = f.Close()
	}()
	assert.NoError(t, err)
	sc := bufio.NewScanner(f)
	var tt table
	for sc.Scan() {
		ff := strings.Fields(sc.Text())
		tt.addRow(newRow(ff...))
	}
	assert.NoError(t, sc.Err())

	return &tt
}
