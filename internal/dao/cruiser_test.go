// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCruiserMeta(t *testing.T) {
	o := loadJSON(t, "crb")

	m := mustMap(o, "metadata")
	assert.Equal(t, "blee", mustField(m, "name"))
}

func TestCruiserSlice(t *testing.T) {
	o := loadJSON(t, "crb")

	s := mustSlice(o, "subjects")
	assert.Equal(t, 1, len(s))
	assert.Equal(t, "fernand", mustField(s[0].(map[string]interface{}), "name"))
	assert.Equal(t, "User", mustField(s[0].(map[string]interface{}), "kind"))
}

// Helpers...

func loadJSON(t assert.TestingT, n string) *unstructured.Unstructured {
	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)

	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)

	return &o
}
