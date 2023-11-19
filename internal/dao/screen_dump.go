// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"os"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*ScreenDump)(nil)
	_ Nuker    = (*ScreenDump)(nil)
)

// ScreenDump represents a scraped resources.
type ScreenDump struct {
	NonResource
}

// Delete a ScreenDump.
func (d *ScreenDump) Delete(_ context.Context, path string, _ *metav1.DeletionPropagation, _ Grace) error {
	return os.Remove(path)
}

// List returns a collection of screen dumps.
func (d *ScreenDump) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	dir, ok := ctx.Value(internal.KeyDir).(string)
	if !ok {
		return nil, errors.New("no screendump dir found in context")
	}

	ff, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	oo := make([]runtime.Object, len(ff))
	for i, f := range ff {
		if fi, err := f.Info(); err == nil {
			oo[i] = render.FileRes{File: fi, Dir: dir}
		}
	}

	return oo, nil
}
