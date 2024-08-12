// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"os"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestScreenDumpRender(t *testing.T) {
	var s render.ScreenDump
	var r model1.Row
	o := render.FileRes{
		File: fileInfo{},
		Dir:  "fred/blee",
	}

	assert.Nil(t, s.Render(o, "fred", &r))
	assert.Equal(t, "fred/blee/bob", r.ID)
	assert.Equal(t, model1.Fields{
		"bob",
		"fred/blee",
		"",
	}, r.Fields[:len(r.Fields)-1])
}

// Helpers...

type fileInfo struct{}

var _ os.FileInfo = fileInfo{}

func (f fileInfo) Name() string       { return "bob" }
func (f fileInfo) Size() int64        { return 100 }
func (f fileInfo) Mode() os.FileMode  { return os.FileMode(0644) }
func (f fileInfo) ModTime() time.Time { return testTime() }
func (f fileInfo) IsDir() bool        { return false }
func (f fileInfo) Sys() interface{}   { return nil }
