// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Len(t, s, 1)
	assert.Equal(t, "fernand", mustField(s[0].(map[string]any), "name"))
	assert.Equal(t, "User", mustField(s[0].(map[string]any), "kind"))
}

// Helpers...

func loadJSON(t require.TestingT, n string) *unstructured.Unstructured {
	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	require.NoError(t, err)

	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	require.NoError(t, err)

	return &o
}
