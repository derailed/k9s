// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
)

// ScreenDump renders a screendumps to screen.
type ScreenDump struct {
	Base
}

// ColorerFunc colors a resource row.
func (ScreenDump) ColorerFunc() model1.ColorerFunc {
	return func(ns string, _ model1.Header, re *model1.RowEvent) tcell.Color {
		return tcell.ColorNavajoWhite
	}
}

// Header returns a header row.
func (ScreenDump) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "DIR"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (b ScreenDump) Render(o interface{}, ns string, r *model1.Row) error {
	f, ok := o.(FileRes)
	if !ok {
		return fmt.Errorf("expecting screendumper, but got %T", o)
	}

	r.ID = filepath.Join(f.Dir, f.File.Name())
	r.Fields = model1.Fields{
		f.File.Name(),
		f.Dir,
		"",
		timeToAge(f.File.ModTime()),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func timeToAge(timestamp time.Time) string {
	return duration.HumanDuration(time.Since(timestamp))
}

// FileRes represents a file resource.
type FileRes struct {
	File os.FileInfo
	Dir  string
}

// GetObjectKind returns a schema object.
func (c FileRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (c FileRes) DeepCopyObject() runtime.Object {
	return c
}
