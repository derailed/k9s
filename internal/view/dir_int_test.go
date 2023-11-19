// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsManifest(t *testing.T) {
	uu := map[string]struct {
		file string
		e    bool
	}{
		"yaml": {file: "fred.yaml", e: true},
		"yml":  {file: "fred.yml", e: true},
		"nope": {file: "fred.txt"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, isManifest(u.file))
		})
	}
}

func TestIsKustomized(t *testing.T) {
	uu := map[string]struct {
		path string
		e    bool
	}{
		"toast": {path: "testdata/fred"},
		"yaml":  {path: "testdata/kmanifests", e: true},
		"yml":   {path: "testdata/k1manifests", e: true},
		"noExt": {path: "testdata/k2manifests", e: true},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, isKustomized(u.path))
		})
	}
}
