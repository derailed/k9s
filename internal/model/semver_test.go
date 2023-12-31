// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewSemVer(t *testing.T) {
	uu := map[string]struct {
		version             string
		major, minor, patch int
	}{
		"plain": {
			version: "0.11.1",
			major:   0,
			minor:   11,
			patch:   1,
		},
		"normalized": {
			version: "v10.11.12",
			major:   10,
			minor:   11,
			patch:   12,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			v := model.NewSemVer(u.version)
			assert.Equal(t, u.major, v.Major)
			assert.Equal(t, u.minor, v.Minor)
			assert.Equal(t, u.patch, v.Patch)
		})
	}
}

func TestSemVerIsCurrent(t *testing.T) {
	uu := map[string]struct {
		current, latest string
		e               bool
	}{
		"same": {
			current: "0.11.1",
			latest:  "0.11.1",
			e:       true,
		},
		"older": {
			current: "v10.11.12",
			latest:  "v10.11.13",
		},
		"newer": {
			current: "10.11.13",
			latest:  "10.11.12",
			e:       true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			v1, v2 := model.NewSemVer(u.current), model.NewSemVer(u.latest)
			assert.Equal(t, u.e, v1.IsCurrent(v2))
		})
	}
}
