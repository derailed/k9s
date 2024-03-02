// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestStatefulSetRender(t *testing.T) {
	c := render.StatefulSet{}
	r := model1.NewRow(4)

	assert.Nil(t, c.Render(load(t, "sts"), "", &r))
	assert.Equal(t, "default/nginx-sts", r.ID)
	assert.Equal(t, model1.Fields{"default", "nginx-sts", "0", "4/4", "app=nginx-sts", "nginx-sts", "nginx", "k8s.gcr.io/nginx-slim:0.8", "app=nginx-sts", ""}, r.Fields[:len(r.Fields)-1])
}
