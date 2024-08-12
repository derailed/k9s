// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog"
	"github.com/sahilm/fuzzy"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestParsePFAnn(t *testing.T) {
	uu := map[string]struct {
		ann, co, port string
		ok            bool
	}{
		"named-port": {
			ann:  "c1:blee",
			co:   "c1",
			port: "blee",
			ok:   true,
		},
		"port-num": {
			ann:  "c1:1234",
			co:   "c1",
			port: "1234",
			ok:   true,
		},
		"toast": {
			ann: "zorg",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			co, port, ok := parsePFAnn(u.ann)
			if u.ok {
				assert.Equal(t, u.co, co)
				assert.Equal(t, u.port, port)
				assert.Equal(t, u.ok, ok)
			}
		})
	}
}

func TestExtractApp(t *testing.T) {
	app := NewApp(mock.NewMockConfig())

	uu := map[string]struct {
		app *App
		err error
	}{
		"cool":     {app: app},
		"not-cool": {err: errors.New("no application found in context")},
	}

	for k := range uu {
		u := uu[k]
		ctx := context.Background()
		if u.app != nil {
			ctx = context.WithValue(ctx, internal.KeyApp, u.app)
		}
		t.Run(k, func(t *testing.T) {
			app, err := extractApp(ctx)
			assert.Equal(t, u.err, err)
			if err == nil {
				assert.Equal(t, u.app, app)
			}
		})
	}
}

func TestFwFQN(t *testing.T) {
	uu := map[string]struct {
		po, co, e string
	}{
		"cool": {po: "p1", co: "c1", e: "p1|c1"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, fwFQN(u.po, u.co))
		})
	}
}

func TestAsKey(t *testing.T) {
	uu := map[string]struct {
		k   string
		err error
		e   tcell.Key
	}{
		"cool": {k: "Ctrl-A", e: tcell.KeyCtrlA},
		"miss": {k: "fred", e: 0, err: errors.New(`invalid key specified: "fred"`)},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			key, err := asKey(u.k)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, key)
		})
	}
}

func TestK8sEnv(t *testing.T) {
	cl, ctx, cfg, u := "cluster1", "context1", "cfg1", "user1"
	flags := genericclioptions.ConfigFlags{
		ClusterName:  &cl,
		Context:      &ctx,
		AuthInfoName: &u,
		KubeConfig:   &cfg,
	}
	c := client.NewConfig(&flags)
	env := k8sEnv(c)

	assert.Equal(t, 5, len(env))
	assert.Equal(t, cl, env["CLUSTER"])
	assert.Equal(t, ctx, env["CONTEXT"])
	assert.Equal(t, u, env["USER"])
	assert.Equal(t, render.NAValue, env["GROUPS"])
	assert.Equal(t, cfg, env["KUBECONFIG"])
}

func TestK9sEnv(t *testing.T) {
	cl, ctx, cfg, u := "cluster1", "context1", "cfg1", "user1"
	flags := genericclioptions.ConfigFlags{
		ClusterName:  &cl,
		Context:      &ctx,
		AuthInfoName: &u,
		KubeConfig:   &cfg,
	}
	c := client.NewConfig(&flags)
	h := model1.Header{
		{Name: "A"},
		{Name: "B"},
		{Name: "C"},
	}
	r := model1.Row{
		Fields: []string{"a1", "b1", "c1"},
	}
	env := defaultEnv(c, "fred/blee", h, &r)

	assert.Equal(t, 10, len(env))
	assert.Equal(t, cl, env["CLUSTER"])
	assert.Equal(t, ctx, env["CONTEXT"])
	assert.Equal(t, u, env["USER"])
	assert.Equal(t, render.NAValue, env["GROUPS"])
	assert.Equal(t, cfg, env["KUBECONFIG"])
	assert.Equal(t, "fred", env["NAMESPACE"])
	assert.Equal(t, "blee", env["NAME"])
	assert.Equal(t, "a1", env["COL-A"])
	assert.Equal(t, "b1", env["COL-B"])
	assert.Equal(t, "c1", env["COL-C"])
}

func TestIsTCPPort(t *testing.T) {
	uu := map[string]struct {
		p string
		e bool
	}{
		"tcp": {"80╱TCP", true},
		"udp": {"80╱UDP", false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, isTCPPort(u.p))
		})
	}
}

func TestFQN(t *testing.T) {
	uu := map[string]struct {
		ns, n, e string
	}{
		"fullFQN": {"blee", "fred", "blee/fred"},
		"allNS":   {"", "fred", "fred"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, fqn(u.ns, u.n))
		})
	}
}

func TestUrlFor(t *testing.T) {
	uu := map[string]struct {
		cfg      config.BenchConfig
		co, port string
		e        string
	}{
		"empty": {
			config.BenchConfig{}, "c1", "9000", "http://localhost:9000/",
		},
		"path": {
			config.BenchConfig{
				HTTP: config.HTTP{
					Path: "/fred/blee",
				},
			},
			"c1",
			"9000",
			"http://localhost:9000/fred/blee",
		},
		"host/path": {
			config.BenchConfig{
				HTTP: config.HTTP{
					Host: "zorg",
					Path: "/fred/blee",
				},
			},
			"c1",
			"9000",
			"http://zorg:9000/fred/blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, urlFor(u.cfg, u.port))
		})
	}
}

func TestContainerID(t *testing.T) {
	uu := map[string]struct {
		path, co string
		e        string
	}{
		"plain": {
			"fred/blee", "c1", "fred/blee:c1",
		},
		"podID": {
			"fred/blee-78f8b5d78c-f8588", "c1", "fred/blee:c1",
		},
		"stsID": {
			"fred/blee-1", "c1", "fred/blee:c1",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, containerID(u.path, u.co))
		})
	}
}

func Test_linesWithRegions(t *testing.T) {
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
				{Index: 1, MatchedIndexes: []int{0, 1, 2}},
			},
			e: []string{"foo", matchTag(0, "bar"), "baz"},
		},
		"single-character": {
			lines: []string{"foo", "bar", "baz"},
			matches: fuzzy.Matches{
				{Index: 1, MatchedIndexes: []int{1}},
			},
			e: []string{"foo", "b" + matchTag(0, "a") + "r", "baz"},
		},
		"multiple-matches": {
			lines: []string{"foo", "bar", "baz"},
			matches: fuzzy.Matches{
				{Index: 1, MatchedIndexes: []int{0, 1, 2}},
				{Index: 2, MatchedIndexes: []int{0, 1, 2}},
			},
			e: []string{"foo", matchTag(0, "bar"), matchTag(1, "baz")},
		},
		"multiple-matches-same-line": {
			lines: []string{"foosfoo baz", "dfbarfoos bar"},
			matches: fuzzy.Matches{
				{Index: 0, MatchedIndexes: []int{0, 1, 2}},
				{Index: 0, MatchedIndexes: []int{4, 5, 6}},
				{Index: 1, MatchedIndexes: []int{5, 6, 7}},
			},
			e: []string{
				matchTag(0, "foo") + "s" + matchTag(1, "foo") + " baz",
				"dfbar" + matchTag(2, "foo") + "s bar",
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, u.e, linesWithRegions(u.lines, u.matches))
		})
	}
}
