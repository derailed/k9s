package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestCronJobRender(t *testing.T) {
	c := render.CronJob{}
	r := render.NewRow(6)
	c.Render(load(t, "cj"), "", &r)

	assert.Equal(t, "default/hello", r.ID)
	assert.Equal(t, render.Fields{"default", "hello", "*/1 * * * *", "false", "0"}, r.Fields[:5])
}
