package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestPersistentVolumeClaimRender(t *testing.T) {
	c := render.PersistentVolumeClaim{}
	r := render.NewRow(8)
	c.Render(load(t, "pvc"), "", &r)

	assert.Equal(t, "default/www-nginx-sts-0", r.ID)
	assert.Equal(t, render.Fields{"default", "www-nginx-sts-0", "Bound", "pvc-fbabd470-8725-11e9-a8e8-42010a80015b", "1Gi", "RWO", "standard"}, r.Fields[:7])
}
