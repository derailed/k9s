// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
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
