package dao

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*ScreenDump)(nil)
	_ Nuker    = (*ScreenDump)(nil)

	// InvalidCharsRX contains invalid filename characters.
	invalidPathCharsRX = regexp.MustCompile(`[:/\\]+`)
)

// ScreenDump represents a scraped resources.
type ScreenDump struct {
	NonResource
}

// Delete a ScreenDump.
func (d *ScreenDump) Delete(path string, cascade, force bool) error {
	return os.Remove(path)
}

// List returns a collection of screen dumps.
func (d *ScreenDump) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	dir, ok := ctx.Value(internal.KeyDir).(string)
	if !ok {
		return nil, errors.New("no screendump dir found in context")
	}

	ff, err := ioutil.ReadDir(SanitizeFilename(dir))
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(ff))
	for i, f := range ff {
		oo[i] = render.FileRes{File: f, Dir: dir}
	}

	return oo, nil
}

// Helpers...

// SanitizeFilename sanitizes the dump filename.
func SanitizeFilename(name string) string {
	return invalidPathCharsRX.ReplaceAllString(name, "-")
}
