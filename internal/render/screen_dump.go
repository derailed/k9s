package render

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell"
)

// ScreenDump renders a screendumps to screen.
type ScreenDump struct{}

// ColorerFunc colors a resource row.
func (ScreenDump) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorNavajoWhite
	}
}

type DecoratorFunc func(string) string

var ageDecorator = func(a string) string {
	return toAgeHuman(a)
}

// Header returns a header row.
func (ScreenDump) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "AGE", Decorator: ageDecorator},
	}
}

// Render renders a K8s resource to screen.
func (b ScreenDump) Render(o interface{}, ns string, r *Row) error {
	f, ok := o.(ScreenDumper)
	if !ok {
		return fmt.Errorf("Expected string, but got %T", o)
	}

	r.ID = filepath.Join(f.GetDir(), f.GetFile().Name())
	r.Fields = Fields{
		f.GetFile().Name(),
		timeToAge(f.GetFile().ModTime()),
	}

	return nil
}

// Helpers...

func timeToAge(timestamp time.Time) string {
	return time.Since(timestamp).String()
}

type ScreenDumper interface {
	GetFile() os.FileInfo
	GetDir() string
}
