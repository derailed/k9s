package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestPersistentVolumeRender(t *testing.T) {
	c := render.PersistentVolume{}
	r := render.NewRow(9)
	c.Render(load(t, "pv"), "-", &r)

	assert.Equal(t, "-/pvc-07aa4e2c-8726-11e9-a8e8-42010a80015b", r.ID)
	assert.Equal(t, render.Fields{"pvc-07aa4e2c-8726-11e9-a8e8-42010a80015b", "1Gi", "RWO", "Delete", "Bound", "default/www-nginx-sts-1", "standard"}, r.Fields[:7])
}

func TestTerminatingPersistentVolumeRender(t *testing.T) {
	c := render.PersistentVolume{}
	r := render.NewRow(9)
	c.Render(load(t, "pv_terminating"), "-", &r)

	assert.Equal(t, "-/pvc-a4d86f51-916c-476b-83af-b551c91a8ac0", r.ID)
	assert.Equal(t, render.Fields{"pvc-a4d86f51-916c-476b-83af-b551c91a8ac0", "1Gi", "RWO", "Delete", "Terminating", "default/www-nginx-sts-2", "standard"}, r.Fields[:7])
}
