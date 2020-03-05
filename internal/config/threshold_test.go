package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDefConFor(t *testing.T) {
	uu := map[string]struct {
		k string
		v int
		e config.DefConLevel
	}{
		"normal": {
			k: "cpu",
			v: 0,

			e: config.DefCon5,
		},
		"4": {
			k: "cpu",
			v: 71,
			e: config.DefCon4,
		},
		"3": {
			k: "cpu",
			v: 75,
			e: config.DefCon3,
		},
		"2": {
			k: "cpu",
			v: 80,
			e: config.DefCon2,
		},
		"1": {
			k: "cpu",
			v: 100,
			e: config.DefCon1,
		},
		"over": {
			k: "cpu",
			v: 150,
			e: config.DefCon5,
		},
	}

	o := config.NewThreshold()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, o.DefConFor(u.k, u.v))
		})
	}
}
