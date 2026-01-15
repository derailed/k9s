// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*ContainerFs)(nil)

// ContainerFs provides access to container filesystem resources.
type ContainerFs struct {
	NonResource
}

// NewContainerFs returns a new container filesystem accessor.
func NewContainerFs(f Factory) *ContainerFs {
	var c ContainerFs
	c.Init(f, client.CfsGVR)
	return &c
}

// List returns a collection of container filesystem entries.
func (*ContainerFs) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	// Extract context values
	podPath, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("no pod path in context")
	}

	container, ok := ctx.Value(internal.KeyContainers).(string)
	if !ok {
		return nil, errors.New("no container in context")
	}

	currentDir, ok := ctx.Value(internal.KeyCurrentDir).(string)
	if !ok {
		currentDir = "/" // Default to root
	}

	// Get mock entries for current directory
	entries, exists := mockFilesystem[currentDir]
	if !exists {
		return nil, fmt.Errorf("directory not found: %s", currentDir)
	}

	// Prevent unused variable warnings
	_, _ = podPath, container

	// Convert to runtime.Object slice
	oo := make([]runtime.Object, 0, len(entries))
	for _, e := range entries {
		fullPath := filepath.Join(currentDir, e.name)
		if currentDir == "/" {
			fullPath = "/" + e.name
		}

		oo = append(oo, render.ContainerFsRes{
			Path:       fullPath,
			Name:       e.name,
			IsDir:      e.isDir,
			Size:       e.size,
			ModTime:    time.Now().Add(e.mod),
			Permission: e.perm,
		})
	}

	return oo, nil
}

// IsDirectory checks if the given path is a directory in the mock filesystem.
func (*ContainerFs) IsDirectory(path string) bool {
	_, exists := mockFilesystem[path]
	return exists
}

// Get fetches a specific resource.
func (*ContainerFs) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, errors.New("get not implemented for container filesystem")
}

// ----------------------------------------------------------------------------
// Mock filesystem data...

type mockEntry struct {
	name  string
	isDir bool
	size  int64
	perm  string
	mod   time.Duration // Relative to now
}

var mockFilesystem = map[string][]mockEntry{
	"/": {
		{name: "bin", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
		{name: "etc", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
		{name: "home", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
		{name: "opt", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
		{name: "tmp", isDir: true, perm: "drwxrwxrwt", mod: -24 * time.Hour},
		{name: "usr", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
		{name: "var", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
	},
	"/bin": {
		{name: "bash", isDir: false, size: 1234567, perm: "-rwxr-xr-x", mod: -48 * time.Hour},
		{name: "sh", isDir: false, size: 128000, perm: "-rwxr-xr-x", mod: -48 * time.Hour},
		{name: "ls", isDir: false, size: 138208, perm: "-rwxr-xr-x", mod: -48 * time.Hour},
		{name: "cat", isDir: false, size: 35064, perm: "-rwxr-xr-x", mod: -48 * time.Hour},
	},
	"/etc": {
		{name: "hosts", isDir: false, size: 256, perm: "-rw-r--r--", mod: -2 * time.Hour},
		{name: "hostname", isDir: false, size: 64, perm: "-rw-r--r--", mod: -2 * time.Hour},
		{name: "resolv.conf", isDir: false, size: 128, perm: "-rw-r--r--", mod: -1 * time.Hour},
		{name: "passwd", isDir: false, size: 1024, perm: "-rw-r--r--", mod: -24 * time.Hour},
		{name: "group", isDir: false, size: 512, perm: "-rw-r--r--", mod: -24 * time.Hour},
		{name: "nginx", isDir: true, perm: "drwxr-xr-x", mod: -5 * time.Hour},
		{name: "ssl", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
	},
	"/etc/nginx": {
		{name: "nginx.conf", isDir: false, size: 2048, perm: "-rw-r--r--", mod: -30 * time.Minute},
		{name: "mime.types", isDir: false, size: 5349, perm: "-rw-r--r--", mod: -48 * time.Hour},
		{name: "conf.d", isDir: true, perm: "drwxr-xr-x", mod: -30 * time.Minute},
		{name: "sites-available", isDir: true, perm: "drwxr-xr-x", mod: -30 * time.Minute},
		{name: "sites-enabled", isDir: true, perm: "drwxr-xr-x", mod: -30 * time.Minute},
	},
	"/etc/nginx/conf.d": {
		{name: "default.conf", isDir: false, size: 1024, perm: "-rw-r--r--", mod: -30 * time.Minute},
	},
	"/etc/nginx/sites-available": {
		{name: "default", isDir: false, size: 2084, perm: "-rw-r--r--", mod: -48 * time.Hour},
	},
	"/etc/nginx/sites-enabled": {
		{name: "default", isDir: false, size: 2084, perm: "-rw-r--r--", mod: -48 * time.Hour},
	},
	"/etc/ssl": {
		{name: "certs", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
		{name: "private", isDir: true, perm: "drwx------", mod: -48 * time.Hour},
	},
	"/etc/ssl/certs": {
		{name: "ca-certificates.crt", isDir: false, size: 214856, perm: "-rw-r--r--", mod: -48 * time.Hour},
	},
	"/etc/ssl/private": {},
	"/home": {
		{name: "app", isDir: true, perm: "drwxr-xr-x", mod: -12 * time.Hour},
	},
	"/home/app": {
		{name: "package.json", isDir: false, size: 512, perm: "-rw-r--r--", mod: -12 * time.Hour},
		{name: "app.js", isDir: false, size: 4096, perm: "-rw-r--r--", mod: -30 * time.Minute},
		{name: "node_modules", isDir: true, perm: "drwxr-xr-x", mod: -12 * time.Hour},
		{name: "public", isDir: true, perm: "drwxr-xr-x", mod: -12 * time.Hour},
	},
	"/home/app/node_modules": {
		{name: "express", isDir: true, perm: "drwxr-xr-x", mod: -12 * time.Hour},
		{name: "dotenv", isDir: true, perm: "drwxr-xr-x", mod: -12 * time.Hour},
	},
	"/home/app/public": {
		{name: "index.html", isDir: false, size: 1024, perm: "-rw-r--r--", mod: -6 * time.Hour},
		{name: "styles.css", isDir: false, size: 2048, perm: "-rw-r--r--", mod: -6 * time.Hour},
	},
	"/opt": {},
	"/tmp": {
		{name: "test.txt", isDir: false, size: 42, perm: "-rw-r--r--", mod: -15 * time.Minute},
	},
	"/usr": {
		{name: "bin", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
		{name: "lib", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
		{name: "local", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
		{name: "share", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
	},
	"/usr/bin": {
		{name: "curl", isDir: false, size: 225792, perm: "-rwxr-xr-x", mod: -48 * time.Hour},
		{name: "wget", isDir: false, size: 519808, perm: "-rwxr-xr-x", mod: -48 * time.Hour},
		{name: "node", isDir: false, size: 36521984, perm: "-rwxr-xr-x", mod: -12 * time.Hour},
		{name: "npm", isDir: false, size: 4096, perm: "-rwxr-xr-x", mod: -12 * time.Hour},
	},
	"/usr/lib": {},
	"/usr/local": {
		{name: "bin", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
		{name: "lib", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
	},
	"/usr/local/bin": {},
	"/usr/local/lib": {},
	"/usr/share": {
		{name: "doc", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
		{name: "man", isDir: true, perm: "drwxr-xr-x", mod: -48 * time.Hour},
	},
	"/usr/share/doc": {},
	"/usr/share/man": {},
	"/var": {
		{name: "log", isDir: true, perm: "drwxr-xr-x", mod: -10 * time.Minute},
		{name: "lib", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
		{name: "cache", isDir: true, perm: "drwxr-xr-x", mod: -24 * time.Hour},
		{name: "run", isDir: true, perm: "drwxr-xr-x", mod: -5 * time.Minute},
	},
	"/var/log": {
		{name: "access.log", isDir: false, size: 1048576, perm: "-rw-r--r--", mod: -5 * time.Minute},
		{name: "error.log", isDir: false, size: 4096, perm: "-rw-r--r--", mod: -1 * time.Hour},
		{name: "nginx", isDir: true, perm: "drwxr-xr-x", mod: -5 * time.Minute},
	},
	"/var/log/nginx": {
		{name: "access.log", isDir: false, size: 2097152, perm: "-rw-r--r--", mod: -5 * time.Minute},
		{name: "error.log", isDir: false, size: 8192, perm: "-rw-r--r--", mod: -1 * time.Hour},
	},
	"/var/lib": {},
	"/var/cache": {},
	"/var/run": {
		{name: "nginx.pid", isDir: false, size: 6, perm: "-rw-r--r--", mod: -5 * time.Minute},
	},
}
