// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Dir)(nil)

// DirLocal tracks standard and custom command aliases.
type DirLocal struct {
	NonResource
}

// NewDirLocal returns a new set of aliases.
func NewDirLocal(f Factory) *Dir {
	var a Dir
	a.Init(f, client.NewGVR("dirlocal"))
	return &a
}

// List returns a collection of aliases.
func (a *DirLocal) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	dir, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("no dir in context")
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, 0, len(files))
	for _, f := range files {
		oo = append(oo, render.DirRes{
			Path:  filepath.Join(dir, f.Name()),
			Entry: f,
		})
	}

	return oo, err
}

// Get fetch a resource.
func (a *DirLocal) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, errors.New("nyi")
}
