package render

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ScreenDump renders a screendumps to screen.
type ScreenDump struct{}

// ColorerFunc colors a resource row.
func (ScreenDump) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorNavajoWhite
	}
}

// DecoratorFunc decorates a string.
type DecoratorFunc func(string) string

// AgeDecorator represents a timestamped as human column.
var AgeDecorator = func(a string) string {
	return toAgeHuman(a)
}

// Header returns a header row.
func (ScreenDump) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "AGE", Decorator: AgeDecorator},
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
