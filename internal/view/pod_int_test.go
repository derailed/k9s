package view

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeShellArgs(t *testing.T) {
	config, empty := "coolConfig", ""
	uu := map[string]struct {
		path, co, context string
		cfg               *string
		e                 string
	}{
		"config": {
			"fred/blee",
			"c1",
			"ctx1",
			&config,
			"exec -it --context ctx1 -n fred blee --kubeconfig coolConfig -c c1 -- sh -c " + shellCheck,
		},
		"noconfig": {
			"fred/blee",
			"c1",
			"ctx1",
			nil,
			"exec -it --context ctx1 -n fred blee -c c1 -- sh -c " + shellCheck,
		},
		"emptyConfig": {
			"fred/blee",
			"c1",
			"ctx1",
			&empty,
			"exec -it --context ctx1 -n fred blee -c c1 -- sh -c " + shellCheck,
		},
		"singleContainer": {
			"fred/blee",
			"",
			"ctx1",
			&empty,
			"exec -it --context ctx1 -n fred blee -- sh -c " + shellCheck,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			args := computeShellArgs(u.path, u.co, u.context, u.cfg)

			assert.Equal(t, u.e, strings.Join(args, " "))
		})
	}
}
