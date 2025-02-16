// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"errors"
	"testing"

	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestCustCol_parse(t *testing.T) {
	uu := map[string]struct {
		s   string
		err error
		e   colDef
	}{
		"empty": {
			err: errors.New(`invalid column definition ""`),
		},

		"plain": {
			s: "fred",
			e: colDef{
				name: "fred",
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignLeft,
				},
			},
		},

		"plain-wide": {
			s: "fred|W",
			e: colDef{
				name: "fred",
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignLeft,
					wide:  true,
				},
			},
		},

		"plain-hide": {
			s: "fred|WH",
			e: colDef{
				name: "fred",
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignLeft,
					wide:  true,
					hide:  true,
				},
			},
		},

		"age": {
			s: "AGE|TR",
			e: colDef{
				name: "AGE",
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignRight,
					time:  true,
				},
			},
		},

		"plain-wide-right": {
			s: "fred|WR",
			e: colDef{
				name: "fred",
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignRight,
					wide:  true,
				},
			},
		},

		"complex": {
			s: "BLEE:.spec.addresses[?(@.type == 'CiliumInternalIP')].ip",
			e: colDef{
				name: "BLEE",
				idx:  -1,
				spec: "{.spec.addresses[?(@.type == 'CiliumInternalIP')].ip}",
				colAttrs: colAttrs{
					align: tview.AlignLeft,
				},
			},
		},

		"complex-wide": {
			s: "BLEE:.spec.addresses[?(@.type == 'CiliumInternalIP')].ip|WR",
			e: colDef{
				name: "BLEE",
				idx:  -1,
				spec: "{.spec.addresses[?(@.type == 'CiliumInternalIP')].ip}",
				colAttrs: colAttrs{
					align: tview.AlignRight,
					wide:  true,
				},
			},
		},

		"full-complex-wide": {
			s: "BLEE:.spec.addresses[?(@.type == 'CiliumInternalIP')].ip|WR",
			e: colDef{
				name: "BLEE",
				idx:  -1,
				spec: "{.spec.addresses[?(@.type == 'CiliumInternalIP')].ip}",
				colAttrs: colAttrs{
					align: tview.AlignRight,
					wide:  true,
				},
			},
		},

		"full-number-wide": {
			s: "fred:.metadata.name|NW",
			e: colDef{
				name: "fred",
				idx:  -1,
				spec: "{.metadata.name}",
				colAttrs: colAttrs{
					align:    tview.AlignRight,
					capacity: true,
					wide:     true,
				},
			},
		},

		"full-wide": {
			s: "fred:.metadata.name|RW",
			e: colDef{
				name: "fred",
				idx:  -1,
				spec: "{.metadata.name}",
				colAttrs: colAttrs{
					align: tview.AlignRight,
					wide:  true,
				},
			},
		},

		"partial-time-no-wide": {
			s: "fred:.metadata.name|T",
			e: colDef{
				name: "fred",
				idx:  -1,
				spec: "{.metadata.name}",
				colAttrs: colAttrs{
					align: tview.AlignLeft,
					time:  true,
				},
			},
		},

		"partial-no-type-no-wide": {
			s: "fred:.metadata.name",
			e: colDef{
				name: "fred",
				idx:  -1,
				spec: "{.metadata.name}",
				colAttrs: colAttrs{
					align: tview.AlignLeft,
				},
			},
		},

		"partial-no-type-wide": {
			s: "fred:.metadata.name|W",
			e: colDef{
				name: "fred",
				idx:  -1,
				spec: "{.metadata.name}",
				colAttrs: colAttrs{
					align: tview.AlignLeft,
					wide:  true,
				},
			},
		},

		"toast": {
			s: "fred||.metadata.name|W",
			e: colDef{
				name: "fred",
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignLeft,
				},
			},
		},

		"toast-no-name": {
			s:   `:.metadata.name.fred|TW`,
			err: errors.New(`invalid column definition ":.metadata.name.fred|TW"`),
		},

		"spec-column-typed": {
			s: `fred:.metadata.name.k8s:fred\.blee|TW`,
			e: colDef{
				name: "fred",
				spec: `{.metadata.name.k8s:fred\.blee}`,
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignLeft,
					time:  true,
					wide:  true,
				},
			},
		},

		"partial-no-spec-no-wide": {
			s: "fred|T",
			e: colDef{
				name: "fred",
				idx:  -1,
				colAttrs: colAttrs{
					align: tview.AlignLeft,
					time:  true,
				},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c, err := parse(u.s)
			if err != nil {
				assert.Equal(t, u.err, err)
			} else {
				assert.Equal(t, u.e, c)
			}
		})
	}
}
