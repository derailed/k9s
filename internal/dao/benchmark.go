package dao

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Benchmark)(nil)
	_ Nuker    = (*Benchmark)(nil)
)

// Benchmark represents a benchmark resource.
type Benchmark struct {
	NonResource
}

// Delete nukes a resource.
func (b *Benchmark) Delete(path string, cascade, force bool) error {
	return os.Remove(path)
}

// Get returns a resource.
func (b *Benchmark) Get(context.Context, string) (runtime.Object, error) {
	panic("NYI")
}

// List returns a collection of resources.
func (b *Benchmark) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	dir, ok := ctx.Value(internal.KeyDir).(string)
	if !ok {
		return nil, errors.New("no benchmark dir found in context")
	}
	path, _ := ctx.Value(internal.KeyPath).(string)

	ff, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, 0, len(ff))
	for _, f := range ff {
		if path != "" && !strings.HasPrefix(f.Name(), strings.Replace(path, "/", "_", 1)) {
			continue
		}
		oo = append(oo, render.BenchInfo{File: f, Path: filepath.Join(dir, f.Name())})
	}

	return oo, nil
}
