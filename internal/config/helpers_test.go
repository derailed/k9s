// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHelperInNSList(t *testing.T) {
	uu := []struct {
		item     string
		list     []interface{}
		expected bool
	}{
		{
			"fred",
			[]interface{}{
				v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fred"}},
			},
			true,
		},
		{
			"blee",
			[]interface{}{
				v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fred"}},
			},
			false,
		},
	}

	for _, u := range uu {
		assert.Equal(t, u.expected, config.InNSList(u.list, u.item))
	}
}
