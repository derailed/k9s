// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*DirRemote)(nil)

// Dir tracks standard and custom command aliases.
type DirRemote struct {
	NonResource
}

// NewDirRemote returns a new set of aliases.
func NewDirRemote(f Factory) *DirRemote {
	var a DirRemote
	a.Init(f, client.NewGVR("dirremote"))
	return &a
}

// List returns a collection of aliases.
func (a *DirRemote) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	dir, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("no dir in context")
	}

	// It would be better to list files here ?
	// No access to view/exec.go, though
	txt, ok := ctx.Value(internal.KeyContents).(string)
	if !ok {
		return nil, errors.New("no contents in context")
	}

	var err error = nil
	lines := strings.Split(txt, "\n")
	oo := make([]runtime.Object, 0, len(lines))
	for _, name := range lines {
		if len(name) == 0 {
			continue
		}
		name = strings.TrimSuffix(name, "\n")
		name = strings.TrimSuffix(name, "\r")
		if name == "./" {
			continue
		}
		if name == "../" && dir == "/" {
			continue
		}
		//if strings.HasSuffix(name, "/") { // directory
		// do not strip the trailing slash
		//name = strings.TrimSuffix(name, "/")
		//}
		name = strings.TrimSuffix(name, "*") // executable
		if strings.HasSuffix(name, "@") {    // symlink
			continue // kubectl cp ignores symlinks
		}
		if strings.HasSuffix(name, "|") { // pipe
			continue
		}
		if strings.HasSuffix(name, "=") { // socket
			continue
		}
		oo = append(oo, render.DirRemoteRes{
			Path: filepath.Join(dir, name),
			Name: name,
		})
	}

	return oo, err
}

// Get fetch a resource.
func (a *DirRemote) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, errors.New("nyi")
}
