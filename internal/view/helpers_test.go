package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
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
