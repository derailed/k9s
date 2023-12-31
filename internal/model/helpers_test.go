// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMetaFQN(t *testing.T) {
	uu := map[string]struct {
		meta metav1.ObjectMeta
		e    string
	}{
		"all_namespaces": {
			meta: metav1.ObjectMeta{Name: "fred"},
			e:    "fred",
		},
		"namespaced": {
			meta: metav1.ObjectMeta{Name: "fred", Namespace: "blee"},
			e:    "blee/fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, model.MetaFQN(u.meta))
		})
	}
}

func TestTruncate(t *testing.T) {
	uu := map[string]struct {
		data string
		size int
		e    string
	}{
		"same": {
			data: "fred",
			size: 4,
			e:    "fred",
		},
		"small": {
			data: "fred",
			size: 10,
			e:    "fred",
		},
		"larger": {
			data: "fred",
			size: 3,
			e:    "fr…",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, model.Truncate(u.data, u.size))
		})
	}
}
