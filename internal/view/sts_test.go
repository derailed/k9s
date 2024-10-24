// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestStatefulSetNew(t *testing.T) {
	s := view.NewStatefulSet(client.NewGVR("apps/v1/statefulsets"))

	assert.Nil(t, s.Init(makeCtx()))
	assert.Equal(t, "StatefulSets", s.Name())
	assert.Equal(t, 14, len(s.Hints()))
}
