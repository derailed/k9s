// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobRender(t *testing.T) {
	c := render.Job{}
	r := model1.NewRow(4)

	require.NoError(t, c.Render(load(t, "job"), "", &r))
	assert.Equal(t, "default/hello-1567179180", r.ID)
	assert.Equal(t, model1.Fields{"default", "hello-1567179180", "n/a", "1/1", "8s", "controller-uid=7473e6d0-cb3b-11e9-990f-42010a800218", "c1", "blang/busybox-bash"}, r.Fields[:8])
}

func TestJobRenderValid(t *testing.T) {
	uu := map[string]struct {
		completions       any
		parallelism       int64
		succeeded, failed int64
		valid             string
	}{
		"happy":                {completions: int64(1), parallelism: 1, succeeded: 1},
		"retried-then-ok":      {completions: int64(1), parallelism: 1, succeeded: 1, failed: 1},
		"failing":              {completions: int64(1), parallelism: 1, failed: 2, valid: "2 pods failed"},
		"queue-done-with-fail": {parallelism: 3, succeeded: 2, failed: 1},
		"queue-failing":        {parallelism: 3, failed: 1, valid: "1 pods failed"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			o := load(t, "job")
			sp := o.Object["spec"].(map[string]any)
			if u.completions == nil {
				delete(sp, "completions")
			} else {
				sp["completions"] = u.completions
			}
			sp["parallelism"] = u.parallelism
			st := o.Object["status"].(map[string]any)
			st["succeeded"] = u.succeeded
			st["failed"] = u.failed

			c := render.Job{}
			r := model1.NewRow(10)
			require.NoError(t, c.Render(o, "", &r))
			assert.Equal(t, u.valid, r.Fields[8])
		})
	}
}
