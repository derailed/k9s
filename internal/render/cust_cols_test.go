// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestParseSpecs(t *testing.T) {
	uu := map[string]struct {
		cols ColsSpecs
		err  error
		e    ColumnSpecs
	}{
		"empty": {
			e: ColumnSpecs{},
		},

		"plain": {
			cols: ColsSpecs{
				"a",
				"b",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"with-spec-plain": {
			cols: ColsSpecs{
				"a",
				"b:.metadata.name",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
					},
					Spec: "{.metadata.name}",
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"with-spec-fq": {
			cols: ColsSpecs{
				"a",
				"b:.metadata.name|NW",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
						Attrs: model1.Attrs{
							Wide:     true,
							Capacity: true,
							Align:    tview.AlignRight,
						},
					},
					Spec: "{.metadata.name}",
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"spec-type-no-wide": {
			cols: ColsSpecs{
				"a",
				"b:.metadata.name|T",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
						Attrs: model1.Attrs{
							Time: true,
						},
					},
					Spec: "{.metadata.name}",
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"plain-wide": {
			cols: ColsSpecs{
				"a",
				"b|W",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name:  "b",
						Attrs: model1.Attrs{Wide: true},
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"no-spec-kind-wide": {
			cols: ColsSpecs{
				"a",
				"b|NW",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
						Attrs: model1.Attrs{
							Align:    tview.AlignRight,
							Capacity: true,
							Wide:     true,
						},
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"toast-spec": {
			cols: ColsSpecs{
				"a",
				"b:{{crap.bozo}}|NW",
				"c",
			},
			err: errors.New(`unexpected path string, expected a 'name1.name2' or '.name1.name2' or '{name1.name2}' or '{.name1.name2}'`),
		},

		"no-spec": {
			cols: ColsSpecs{
				"a",
				"b|NW",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name:  "b",
						Attrs: model1.Attrs{Align: tview.AlignRight, Capacity: true, Wide: true},
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cols, err := u.cols.parseSpecs()
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, cols)
		})
	}
}
