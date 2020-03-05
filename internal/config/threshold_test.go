package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDefConValidate(t *testing.T) {
	uu := map[string]struct {
		d, e *config.DefCon
	}{
		"default": {
			d: config.NewDefCon(),
			e: config.NewDefCon(),
		},
		"toast": {
			d: &config.DefCon{Levels: []int{10}},
			e: config.NewDefCon(),
		},
		"negative": {
			d: &config.DefCon{Levels: []int{-1, 10, 10, 10}},
			e: &config.DefCon{Levels: []int{90, 10, 10, 10}},
		},
		"out-of-range": {
			d: &config.DefCon{Levels: []int{150, 200, 10, 300}},
			e: &config.DefCon{Levels: []int{90, 80, 10, 70}},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.d.Validate()
			assert.Equal(t, u.e, u.d)
		})
	}
}

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
