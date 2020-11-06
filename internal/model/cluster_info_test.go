package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"vbom.ml/util/sortorder"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestVersionCheck(t *testing.T) {
	uu := map[string]struct {
		current, latest string
		e               bool
	}{
		"same": {
			current: "v0.11.1",
			latest:  "v0.11.1",
		},
		"updated": {
			current: "v0.11.1",
			latest:  "v0.12.1",
			e:       true,
		},
		"current": {
			current: "v0.11.1",
			latest:  "v0.09.2",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, sortorder.NaturalLess(u.current, u.latest))
		})
	}
}

func TestClusterMetaDelta(t *testing.T) {
	uu := map[string]struct {
		o, n model.ClusterMeta
		e    bool
	}{
		"empty": {
			o: model.NewClusterMeta(),
			n: model.NewClusterMeta(),
		},
		"same": {
			o: makeClusterMeta("fred"),
			n: makeClusterMeta("fred"),
		},
		"diff": {
			o: makeClusterMeta("fred"),
			n: makeClusterMeta("freddie"),
			e: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.o.Deltas(u.n))
		})
	}
}

// Helpers...

func makeClusterMeta(cluster string) model.ClusterMeta {
	m := model.NewClusterMeta()
	m.Cluster = cluster
	m.Cpu, m.Mem = 10, 20

	return m
}
