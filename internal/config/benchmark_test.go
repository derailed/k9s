// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBenchEmpty(t *testing.T) {
	uu := map[string]struct {
		b Benchmark
		e bool
	}{
		"empty":    {Benchmark{}, true},
		"notEmpty": {newBenchmark(), false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.b.Empty())
		})
	}
}

func TestBenchLoad(t *testing.T) {
	uu := map[string]struct {
		file     string
		c, n     int
		svcCount int
		coCount  int
	}{
		"goodConfig": {
			"testdata/benchmarks/b_good.yaml",
			2,
			1000,
			2,
			0,
		},
		"malformed": {
			"testdata/benchmarks/b_toast.yaml",
			1,
			200,
			0,
			0,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			b, err := NewBench(u.file)

			require.NoError(t, err)
			assert.Equal(t, u.c, b.Benchmarks.Defaults.C)
			assert.Equal(t, u.n, b.Benchmarks.Defaults.N)
			assert.Len(t, b.Benchmarks.Services, u.svcCount)
			assert.Len(t, b.Benchmarks.Containers, u.coCount)
		})
	}
}

func TestBenchServiceLoad(t *testing.T) {
	uu := map[string]struct {
		key                string
		c, n               int
		method, host, path string
		http2              bool
		body               string
		auth               Auth
		headers            http.Header
	}{
		"s1": {
			"default/nginx",
			2,
			1000,
			"GET",
			"10.10.10.10",
			"/",
			true,
			`{"fred": "blee"}`,
			Auth{"fred", "blee"},
			http.Header{"Accept": []string{"text/html"}, "Content-Type": []string{"application/json"}},
		},
		"s2": {
			"blee/fred",
			10,
			1500,
			"POST",
			"20.20.20.20",
			"/zorg",
			false,
			`{"fred": "blee"}`,
			Auth{"fred", "blee"},
			http.Header{"Accept": []string{"text/html"}, "Content-Type": []string{"application/json"}},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			b, err := NewBench("testdata/benchmarks/b_good.yaml")

			require.NoError(t, err)
			assert.Len(t, b.Benchmarks.Services, 2)
			svc := b.Benchmarks.Services[u.key]
			assert.Equal(t, u.c, svc.C)
			assert.Equal(t, u.n, svc.N)
			assert.Equal(t, u.method, svc.HTTP.Method)
			assert.Equal(t, u.host, svc.HTTP.Host)
			assert.Equal(t, u.path, svc.HTTP.Path)
			assert.Equal(t, u.http2, svc.HTTP.HTTP2)
			assert.Equal(t, u.body, svc.HTTP.Body)
			assert.Equal(t, u.auth, svc.Auth)
			assert.Equal(t, u.headers, svc.HTTP.Headers)
		})
	}
}

func TestBenchReLoad(t *testing.T) {
	b, err := NewBench("testdata/benchmarks/b_containers.yaml")
	require.NoError(t, err)

	assert.Equal(t, 2, b.Benchmarks.Defaults.C)
	require.NoError(t, b.Reload("testdata/benchmarks/b_containers_1.yaml"))
	assert.Equal(t, 20, b.Benchmarks.Defaults.C)
}

func TestBenchLoadToast(t *testing.T) {
	_, err := NewBench("testdata/toast.yaml")
	assert.Error(t, err)
}

func TestBenchContainerLoad(t *testing.T) {
	uu := map[string]struct {
		key                string
		c, n               int
		method, host, path string
		http2              bool
		body               string
		auth               Auth
		headers            http.Header
	}{
		"c1": {
			"c1",
			2,
			1000,
			"GET",
			"10.10.10.10",
			"/duh",
			true,
			`{"fred": "blee"}`,
			Auth{"fred", "blee"},
			http.Header{"Accept": []string{"text/html"}, "Content-Type": []string{"application/json"}},
		},
		"c2": {
			"c2",
			10,
			1500,
			"POST",
			"20.20.20.20",
			"/fred",
			false,
			`{"fred": "blee"}`,
			Auth{"fred", "blee"},
			http.Header{"Accept": []string{"text/html"}, "Content-Type": []string{"application/json"}},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			b, err := NewBench("testdata/benchmarks/b_containers.yaml")

			require.NoError(t, err)
			assert.Len(t, b.Benchmarks.Services, 2)
			co := b.Benchmarks.Containers[u.key]
			assert.Equal(t, u.c, co.C)
			assert.Equal(t, u.n, co.N)
			assert.Equal(t, u.method, co.HTTP.Method)
			assert.Equal(t, u.host, co.HTTP.Host)
			assert.Equal(t, u.path, co.HTTP.Path)
			assert.Equal(t, u.http2, co.HTTP.HTTP2)
			assert.Equal(t, u.body, co.HTTP.Body)
			assert.Equal(t, u.auth, co.Auth)
			assert.Equal(t, u.headers, co.HTTP.Headers)
		})
	}
}
