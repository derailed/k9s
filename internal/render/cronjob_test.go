// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestCronJobRender(t *testing.T) {
	c := render.CronJob{}
	r := model1.NewRow(6)

	assert.NoError(t, c.Render(load(t, "cj"), "", &r))
	assert.Equal(t, "default/hello", r.ID)
	assert.Equal(t, model1.Fields{"default", "hello", "0", "*/1 * * * *", "false", "0"}, r.Fields[:6])
}
