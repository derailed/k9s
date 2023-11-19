// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestLogItemsFilter(t *testing.T) {
	uu := map[string]struct {
		q    string
		opts dao.LogOptions
		e    []int
		err  error
	}{
		"empty": {
			opts: dao.LogOptions{},
		},
		"pod-name": {
			q: "blee",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{0, 1, 2},
		},
		"container-name": {
			q: "c1",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{0, 1, 2},
		},
		"message": {
			q: "zorg",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{2},
		},
		"fuzzy": {
			q: "-f zorg",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{2},
		},
	}

	for k := range uu {
		u := uu[k]
		ii := dao.NewLogItems()
		ii.Add(
			dao.NewLogItem([]byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))),
			dao.NewLogItemFromString("Bumble bee tuna"),
			dao.NewLogItemFromString("Jean Batiste Emmanuel Zorg"),
		)
		t.Run(k, func(t *testing.T) {
			_, n := client.Namespaced(u.opts.Path)
			for _, i := range ii.Items() {
				i.Pod, i.Container = n, u.opts.Container
			}
			res, _, err := ii.Filter(0, u.q, false)
			assert.Equal(t, u.err, err)
			if err == nil {
				assert.Equal(t, u.e, res)
			}
		})
	}
}

func TestLogItemsRender(t *testing.T) {
	uu := map[string]struct {
		opts dao.LogOptions
		e    string
	}{
		"empty": {
			opts: dao.LogOptions{},
			e:    "Testing 1,2,3...\n",
		},
		"container": {
			opts: dao.LogOptions{
				Container: "fred",
			},
			e: "[teal::b]fred[-::-] Testing 1,2,3...\n",
		},
		"pod-container": {
			opts: dao.LogOptions{
				Path:      "blee/fred",
				Container: "blee",
			},
			e: "[teal::]fred [teal::b]blee[-::-] Testing 1,2,3...\n",
		},
		"full": {
			opts: dao.LogOptions{
				Path:          "blee/fred",
				Container:     "blee",
				ShowTimestamp: true,
			},
			e: "[gray::b]2018-12-14T10:36:43.326972-07:00 [-::-][teal::]fred [teal::b]blee[-::-] Testing 1,2,3...\n",
		},
	}

	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	for k := range uu {
		ii := dao.NewLogItems()
		ii.Add(dao.NewLogItem(s))
		u := uu[k]
		_, n := client.Namespaced(u.opts.Path)
		ii.Items()[0].Pod, ii.Items()[0].Container = n, u.opts.Container
		t.Run(k, func(t *testing.T) {
			res := make([][]byte, 1)
			ii.Render(0, u.opts.ShowTimestamp, res)
			assert.Equal(t, u.e, string(res[0]))
		})
	}
}
