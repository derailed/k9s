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
