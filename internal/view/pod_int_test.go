// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeShellArgs(t *testing.T) {
	config, empty := "coolConfig", ""
	_ = config
	uu := map[string]struct {
		fqn, co, os string
		cfg         *string
		e           string
	}{
		"config": {
			"fred/blee",
			"c1",
			"darwin",
			&config,
			"exec -it -n fred blee --kubeconfig coolConfig -c c1 -- sh -c " + shellCheck,
		},
		"no-config": {
			"fred/blee",
			"c1",
			"linux",
			nil,
			"exec -it -n fred blee -c c1 -- sh -c " + shellCheck,
		},
		"empty-config": {
			"fred/blee",
			"",
			"",
			&empty,
			"exec -it -n fred blee -- sh -c " + shellCheck,
		},
		"single-container": {
			"fred/blee",
			"",
			"linux",
			&empty,
			"exec -it -n fred blee -- sh -c " + shellCheck,
		},
		"windows": {
			"fred/blee",
			"c1",
			windowsOS,
			&empty,
			"exec -it -n fred blee -c c1 -- powershell",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			args := computeShellArgs(u.fqn, u.co, u.cfg, u.os)
			assert.Equal(t, u.e, strings.Join(args, " "))
		})
	}
}

// func TestComputeShellArgs(t *testing.T) {
// 	config, empty := "coolConfig", ""
// 	uu := map[string]struct {
// 		path, co string
// 		cfg      *string
// 		e        string
// 	}{
// 		"config": {
// 			"fred/blee",
// 			"c1",
// 			&config,
// 			"exec -it -n fred blee --kubeconfig coolConfig -c c1 -- sh -c " + shellCheck,
// 		},
// 		"noconfig": {
// 			"fred/blee",
// 			"c1",
// 			nil,
// 			"exec -it -n fred blee -c c1 -- sh -c " + shellCheck,
// 		},
// 		"emptyConfig": {
// 			"fred/blee",
// 			"c1",
// 			&empty,
// 			"exec -it -n fred blee -c c1 -- sh -c " + shellCheck,
// 		},
// 		"singleContainer": {
// 			"fred/blee",
// 			"",
// 			&empty,
// 			"exec -it -n fred blee -- sh -c " + shellCheck,
// 		},
// 	}

// 	for k := range uu {
// 		u := uu[k]
// 		t.Run(k, func(t *testing.T) {
// 			args := computeShellArgs(u.path, u.co, u.cfg)

// 			assert.Equal(t, u.e, strings.Join(args, " "))
// 		})
// 	}
// }
