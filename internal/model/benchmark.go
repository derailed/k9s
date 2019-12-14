package model

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

// Benchmark represents a collection of benchmarks.
type Benchmark struct {
	Resource
}

// List returns a collection of screen dumps.
func (b *Benchmark) List(ctx context.Context) ([]runtime.Object, error) {
	dir, ok := ctx.Value(internal.KeyDir).(string)
	if !ok {
		return nil, errors.New("no benchmark dir found in context")
	}

	ff, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(ff))
	for i, f := range ff {
		oo[i] = render.BenchInfo{File: f, Path: filepath.Join(dir, f.Name())}
	}

	return oo, nil
}

// Hydrate returns a pod as container rows.
func (b *Benchmark) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		if err := re.Render(o, render.NonResource, &rr[i]); err != nil {
			return err
		}
	}
	return nil
}
