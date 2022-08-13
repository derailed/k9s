package render

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ScreenDump renders a screendumps to screen.
type ScreenDump struct {
	Base
}

// ColorerFunc colors a resource row.
func (ScreenDump) ColorerFunc() ColorerFunc {
	return func(ns string, _ Header, re RowEvent) tcell.Color {
		return tcell.ColorNavajoWhite
	}
}

// Header returns a header row.
func (ScreenDump) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "DIR"},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (b ScreenDump) Render(o interface{}, ns string, r *Row) error {
	f, ok := o.(FileRes)
	if !ok {
		return fmt.Errorf("expecting screendumper, but got %T", o)
	}

	r.ID = filepath.Join(f.Dir, f.File.Name())
	r.Fields = Fields{
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
	return time.Since(timestamp).String()
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
