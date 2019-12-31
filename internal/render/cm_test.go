package render_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCMRender(t *testing.T) {
	c := render.ConfigMap{}
	r := render.NewRow(4)
	c.Render(load(t, "cm"), "", &r)

	assert.Equal(t, "default/blee", r.ID)
	assert.Equal(t, render.Fields{"default", "blee", "2"}, r.Fields[:3])
}

// Helpers...

func load(t *testing.T, n string) *unstructured.Unstructured {
	raw, err := ioutil.ReadFile(fmt.Sprintf("assets/%s.json", n))
	assert.Nil(t, err)

	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)

	return &o
}
