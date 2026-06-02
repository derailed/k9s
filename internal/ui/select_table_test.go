// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"

	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestClearMark(t *testing.T) {
	uu := map[string]struct {
		marks []string
	}{
		"empty": {},

		"single": {
			marks: []string{"id1"},
		},

		"multiple": {
			marks: []string{"id1", "id2", "id3"},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			s := SelectTable{
				Table: tview.NewTable(),
				marks: sets.New[string](),
			}
			for _, id := range u.marks {
				s.marks.Insert(id)
			}
			s.ClearMarks()
			assert.Empty(t, s.marks)
		})
	}
}

func TestDeleteMark(t *testing.T) {
	uu := map[string]struct {
		marks   []string
		deletes []string
		e       []string
	}{
		"empty": {
			deletes: []string{"foo"},
			e:       []string{},
		},

		"delete-none": {
			marks:   []string{"id1", "id2", "id3"},
			deletes: []string{"id5", "id6", "id4"},
			e:       []string{"id1", "id2", "id3"},
		},

		"delete-multiple": {
			marks:   []string{"id1", "id2", "id3"},
			deletes: []string{"id1", "id2"},
			e:       []string{"id3"},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			s := SelectTable{
				Table: tview.NewTable(),
				marks: sets.New[string](),
			}
			for _, id := range u.marks {
				s.marks.Insert(id)
			}
			for _, id := range u.deletes {
				s.marks.Delete(id)
			}
			assert.Equal(t, u.e, sets.List(s.marks))
		})
	}
}
