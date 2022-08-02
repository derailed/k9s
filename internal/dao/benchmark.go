package dao

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Benchmark)(nil)
	_ Nuker    = (*Benchmark)(nil)

	BenchRx = regexp.MustCompile(`[:|]+`)
)

// Benchmark represents a benchmark resource.
type Benchmark struct {
	NonResource
}

// Delete nukes a resource.
func (b *Benchmark) Delete(_ context.Context, path string, _ *metav1.DeletionPropagation, force bool) error {
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

	ff, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fileName := BenchRx.ReplaceAllString(strings.Replace(path, "/", "_", 1), "_")
	oo := make([]runtime.Object, 0, len(ff))
	for _, f := range ff {
		if path != "" && !strings.HasPrefix(f.Name(), fileName) {
			continue
		}

		if fi, err := f.Info(); err == nil {
			oo = append(oo, render.BenchInfo{File: fi, Path: filepath.Join(dir, f.Name())})
		}
	}

	return oo, nil
}
