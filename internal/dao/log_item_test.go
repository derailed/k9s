// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
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
		"json-off": {
			opts: dao.LogOptions{
				Path: "blee/fred",
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", `{"level":"info","tags":["a","b"]}`),
			e:   `[yellow::]fred[-::] {"level":"info","tags":["a","b"[]}` + "\n",
		},
		"json-pretty": {
			opts: dao.LogOptions{
				Path:       "blee/fred",
				PrettyJSON: true,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", `{"level":"info","tags":["a","b"]}`),
			e:   "[yellow::]fred[-::] {\n  \"level\": \"info\",\n  \"tags\": [\n    \"a\",\n    \"b\"\n  ]\n}\n",
		},
		"json-pretty-timestamp": {
			opts: dao.LogOptions{
				Container:     "blee",
				ShowTimestamp: true,
				PrettyJSON:    true,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", `{"a":1}`),
			e:   "[gray::b]2018-12-14T10:36:43.326972-07:00 [-::-][yellow::b]blee[-::-] {\n  \"a\": 1\n}\n",
		},
		"json-pretty-escape": {
			opts: dao.LogOptions{
				PrettyJSON: true,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", `{"msg":"[red]boom"}`),
			e:   "{\n  \"msg\": \"[red[]boom\"\n}\n",
		},
		"json-pretty-not-json": {
			opts: dao.LogOptions{
				Path:       "blee/fred",
				PrettyJSON: true,
			},
			log: fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "2021-10-28T13:06:37Z [INFO] [blah-blah] Testing 1,2,3..."),
			e:   "[yellow::]fred[-::] 2021-10-28T13:06:37Z [INFO[] [blah-blah[] Testing 1,2,3...\n",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			i := dao.NewRawLogItem([]byte(u.log))
			_, n := client.Namespaced(u.opts.Path)
			i.Pod, i.Container = n, u.opts.Container

			bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
			i.Render("yellow", u.opts.RenderOpts(), bb)
			assert.Equal(t, u.e, bb.String())
		})
	}
}

func TestLogItemRenderError(t *testing.T) {
	var opts dao.LogOptions
	i := opts.ToErrLogItem(errors.New("boom"))

	bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
	i.Render("yellow", dao.RenderOpts{PrettyJSON: true}, bb)
	assert.Equal(t, "[orange::b]boom[::-]\n", bb.String())
}

func BenchmarkLogItemRenderTS(b *testing.B) {
	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	i := dao.NewLogItem(s)
	i.Pod, i.Container = "fred", "blee"

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
		i.Render("yellow", dao.RenderOpts{ShowTime: true}, bb)
	}
}

func BenchmarkLogItemRenderNoTS(b *testing.B) {
	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	i := dao.NewLogItem(s)
	i.Pod, i.Container = "fred", "blee"

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
		i.Render("yellow", dao.RenderOpts{}, bb)
	}
}

func BenchmarkLogItemRenderPrettyJSON(b *testing.B) {
	const ts = "2018-12-14T10:36:43.326972-07:00 "
	uu := []struct {
		name string
		line string
		opts dao.RenderOpts
	}{
		{"plain/off", ts + "[INFO] request served GET /api/v1/users status=200 latency=12.5ms\n", dao.RenderOpts{}},
		{"plain/on", ts + "[INFO] request served GET /api/v1/users status=200 latency=12.5ms\n", dao.RenderOpts{PrettyJSON: true}},
		{"json/off", ts + `{"level":"info","msg":"request served","method":"GET","path":"/api/v1/users","status":200,"latency_ms":12.5,"tags":["a","b"]}` + "\n", dao.RenderOpts{}},
		{"json/on", ts + `{"level":"info","msg":"request served","method":"GET","path":"/api/v1/users","status":200,"latency_ms":12.5,"tags":["a","b"]}` + "\n", dao.RenderOpts{PrettyJSON: true}},
	}

	for _, u := range uu {
		b.Run(u.name, func(b *testing.B) {
			i := dao.NewRawLogItem([]byte(u.line))
			i.Pod, i.Container = "fred", "blee"

			b.ReportAllocs()
			for range b.N {
				bb := bytes.NewBuffer(make([]byte, 0, i.Size()))
				i.Render("yellow", u.opts, bb)
			}
		})
	}
}
