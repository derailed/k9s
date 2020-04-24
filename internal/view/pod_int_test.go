package view

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeShellArgs(t *testing.T) {
	config, empty := "coolConfig", ""
	uu := map[string]struct {
		path, co string
		cfg      *string
		e        string
	}{
		"config": {
			"fred/blee",
			"c1",
			&config,
			"exec -it -n fred blee --kubeconfig coolConfig -c c1 -- sh -c " + shellCheck,
		},
		"noconfig": {
			"fred/blee",
			"c1",
			nil,
			"exec -it -n fred blee -c c1 -- sh -c " + shellCheck,
		},
		"emptyConfig": {
			"fred/blee",
			"c1",
			&empty,
			"exec -it -n fred blee -c c1 -- sh -c " + shellCheck,
		},
		"singleContainer": {
			"fred/blee",
			"",
			&empty,
			"exec -it -n fred blee -- sh -c " + shellCheck,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			args := computeShellArgs(u.path, u.co, u.cfg)

			assert.Equal(t, u.e, strings.Join(args, " "))
		})
	}
}
