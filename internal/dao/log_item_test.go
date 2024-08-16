// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestLogItemEmpty(t *testing.T) {
	uu := map[string]struct {
		s string
		e bool
	}{
		"empty": {s: "", e: true},
		"full":  {s: "Testing 1,2,3..."},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			i := dao.NewLogItemFromString(u.s)
			assert.Equal(t, u.e, i.IsEmpty())
		})
	}
}

func TestLogItemRender(t *testing.T) {
	uu := map[string]struct {
		opts dao.LogOptions
		log  string
		e    string
	}{
		"empty": {
			opts: dao.LogOptions{},
			log:  fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."),
			e:    "Testing 1,2,3...\n",
		},
		"container": {
			opts: dao.LogOptions{
				Container: "fred",
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."),
			e:   "[yellow::b]fred[-::-] Testing 1,2,3...\n",
		},
		"pod": {
			opts: dao.LogOptions{
				Path:            "blee/fred",
				Container:       "blee",
				SingleContainer: true,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."),
			e:   "[yellow::]fred [yellow::b]blee[-::-] Testing 1,2,3...\n",
		},
		"full": {
			opts: dao.LogOptions{
				Path:            "blee/fred",
				Container:       "blee",
				SingleContainer: true,
				ShowTimestamp:   true,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."),
			e:   "[gray::b]2018-12-14T10:36:43.326972-07:00 [-::-][yellow::]fred [yellow::b]blee[-::-] Testing 1,2,3...\n",
		},
		"log-level": {
			opts: dao.LogOptions{
				Path:            "blee/fred",
				Container:       "",
				SingleContainer: false,
				ShowTimestamp:   false,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "2021-10-28T13:06:37Z [INFO] [blah-blah] Testing 1,2,3..."),
			e:   "[yellow::]fred[-::] 2021-10-28T13:06:37Z [INFO[] [blah-blah[] Testing 1,2,3...\n",
		},
		"escape": {
			opts: dao.LogOptions{
				Path:            "blee/fred",
				Container:       "",
				SingleContainer: false,
				ShowTimestamp:   false,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", `{"foo":["bar"]} Server listening on: [::]:5000`),
			e:   `[yellow::]fred[-::] {"foo":["bar"[]} Server listening on: [::[]:5000` + "\n",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			i := dao.NewLogItem([]byte(tview.Escape(u.log)))
			_, n := client.Namespaced(u.opts.Path)
			i.Pod, i.Container = n, u.opts.Container

			bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
			i.Render("yellow", u.opts.ShowTimestamp, bb)
			assert.Equal(t, u.e, bb.String())
		})
	}
}

func BenchmarkLogItemRenderTS(b *testing.B) {
	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	i := dao.NewLogItem(s)
	i.Pod, i.Container = "fred", "blee"

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
		i.Render("yellow", true, bb)
	}
}

func BenchmarkLogItemRenderNoTS(b *testing.B) {
	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	i := dao.NewLogItem(s)
	i.Pod, i.Container = "fred", "blee"

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
		i.Render("yellow", false, bb)
	}
}
