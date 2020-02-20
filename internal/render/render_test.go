package render_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Helpers...

func load(t assert.TestingT, n string) *unstructured.Unstructured {
	raw, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)

	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)

	return &o
}
