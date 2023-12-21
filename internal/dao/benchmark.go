// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

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
func (b *Benchmark) Delete(_ context.Context, path string, _ *metav1.DeletionPropagation, _ Grace) error {
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
	path, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("no path specified in context")
	}
	pathMatch := BenchRx.ReplaceAllString(strings.Replace(path, "/", "_", 1), "_")

	ff, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	oo := make([]runtime.Object, 0, len(ff))
	for _, f := range ff {
		if !strings.HasPrefix(f.Name(), pathMatch) {
			continue
		}
		if fi, err := f.Info(); err == nil {
			oo = append(oo, render.BenchInfo{File: fi, Path: filepath.Join(dir, f.Name())})
		}
	}

	return oo, nil
}
