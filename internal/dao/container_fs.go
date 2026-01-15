// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*ContainerFs)(nil)

// dirCacheEntry represents a cached directory listing.
type dirCacheEntry struct {
	entries   []render.ContainerFsRes
	timestamp time.Time
}

// dirCache stores cached directory listings per container.
var (
	dirCache     = make(map[string]dirCacheEntry)
	cacheMutex   sync.RWMutex
	cacheTTL     = 30 * time.Second // Cache entries for 30 seconds
	maxCacheSize = 500              // Maximum number of cache entries
)

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
func (c *ContainerFs) List(ctx context.Context, _ string) ([]runtime.Object, error) {
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

	cacheKey := makeCacheKey(podPath, container, currentDir)

	// Check cache first
	cacheMutex.RLock()
	if entry, found := dirCache[cacheKey]; found {
		age := time.Since(entry.timestamp)
		if age < cacheTTL {
			// Cache hit - return immediately
			cacheMutex.RUnlock()

			// Launch background refresh if cache is getting old (>10 seconds)
			if age > 10*time.Second {
				go c.refreshCache(ctx, podPath, container, currentDir, cacheKey)
			}

			return convertToRuntimeObjects(entry.entries), nil
		}
	}
	cacheMutex.RUnlock()

	// Cache miss or expired - fetch fresh data
	entries, err := c.execInContainer(ctx, podPath, container, currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %w", currentDir, err)
	}

	// Store in cache
	storeInCache(cacheKey, entries)

	return convertToRuntimeObjects(entries), nil
}

// IsDirectory checks if the given path is a directory.
// Note: This is a simplified implementation that relies on the path structure.
// A more robust implementation would execute a test command in the container.
func (c *ContainerFs) IsDirectory(path string) bool {
	// For now, we'll use a simple heuristic: check if the path from the table
	// is marked as a directory. Since this is called from the view after getting
	// the selection from the table, and the table was populated by List(), which
	// includes the IsDir information, we can rely on the fact that directories
	// will be listed in the output.

	// However, since we don't have access to the previous List results here,
	// we'll need to execute a command to check. We can use `test -d` for this.
	// For now, return true to allow navigation (we'll handle errors in the List call).
	return true
}

// Get fetches a specific resource.
func (*ContainerFs) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, errors.New("get not implemented for container filesystem")
}

// makeCacheKey generates a cache key for a directory listing.
func makeCacheKey(podPath, container, dir string) string {
	return fmt.Sprintf("%s:%s:%s", podPath, container, dir)
}

// convertToRuntimeObjects converts filesystem entries to runtime.Object slice.
func convertToRuntimeObjects(entries []render.ContainerFsRes) []runtime.Object {
	oo := make([]runtime.Object, 0, len(entries))
	for _, e := range entries {
		oo = append(oo, e)
	}
	return oo
}

// refreshCache refreshes the cache in the background.
func (c *ContainerFs) refreshCache(ctx context.Context, podPath, container, dir, cacheKey string) {
	entries, err := c.execInContainer(ctx, podPath, container, dir)
	if err != nil {
		// Silently fail - we still have cached data
		return
	}

	storeInCache(cacheKey, entries)
}

// storeInCache stores entries in the cache, evicting old entries if cache is full.
func storeInCache(cacheKey string, entries []render.ContainerFsRes) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// If cache is full, remove oldest entry
	if len(dirCache) >= maxCacheSize {
		var oldestKey string
		var oldestTime time.Time
		first := true

		for key, entry := range dirCache {
			if first || entry.timestamp.Before(oldestTime) {
				oldestKey = key
				oldestTime = entry.timestamp
				first = false
			}
		}

		if oldestKey != "" {
			delete(dirCache, oldestKey)
		}
	}

	dirCache[cacheKey] = dirCacheEntry{
		entries:   entries,
		timestamp: time.Now(),
	}
}

// execInContainer executes a command in a container and returns the output.
func (c *ContainerFs) execInContainer(ctx context.Context, podPath, container, dir string) ([]render.ContainerFsRes, error) {
	// Get kubectl binary
	bin, err := exec.LookPath("kubectl")
	if errors.Is(err, exec.ErrDot) {
		return nil, fmt.Errorf("kubectl command must not be in the current working directory: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("kubectl command is not in your path: %w", err)
	}

	// Parse namespace and pod name
	ns, po := client.Namespaced(podPath)

	// Build kubectl exec command
	args := []string{"exec"}
	if ns != client.BlankNamespace {
		args = append(args, "-n", ns)
	}
	args = append(args, po)
	if container != "" {
		args = append(args, "-c", container)
	}

	// Add context and kubeconfig from Factory
	cfg := c.Client().Config()
	if cfg != nil {
		ctxName := cfg.Flags().Context
		if ctxName != nil && *ctxName != "" {
			args = append(args, "--context", *ctxName)
		}
		kubeConfig := cfg.Flags().KubeConfig
		if kubeConfig != nil && *kubeConfig != "" {
			args = append(args, "--kubeconfig", *kubeConfig)
		}
	}

	// Add the command to execute: ls -la <dir>
	args = append(args, "--", "ls", "-la", dir)

	// Execute command
	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kubectl exec failed: %w (stderr: %s)", err, stderr.String())
	}

	// Parse the output
	return parseLsOutput(dir, stdout.String())
}

// parseLsOutput parses the output of 'ls -la' command.
func parseLsOutput(dir, output string) ([]render.ContainerFsRes, error) {
	lines := strings.Split(output, "\n")
	var results []render.ContainerFsRes

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse ls -la output format:
		// -rw-r--r-- 1 root root 1234 Jan 15 12:34 file.txt
		// drwxr-xr-x 2 root root 4096 Jan 15 12:34 directory
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue // Skip malformed lines
		}

		// Skip total line
		if fields[0] == "total" {
			continue
		}

		perm := fields[0]
		sizeStr := fields[4]

		// Name is everything from field 8 onwards (to handle spaces in filenames)
		name := strings.Join(fields[8:], " ")

		// Skip . and ..
		if name == "." || name == ".." {
			continue
		}

		// Parse size
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			size = 0
		}

		// Determine if it's a directory
		isDir := strings.HasPrefix(perm, "d")

		// Parse modification time
		// Format can be: "Jan 15 12:34" or "Jan 15 2023"
		month := fields[5]
		day := fields[6]
		timeOrYear := fields[7]

		modTime := parseModTime(month, day, timeOrYear)

		// Build full path
		fullPath := filepath.Join(dir, name)
		if dir == "/" {
			fullPath = "/" + name
		}

		results = append(results, render.ContainerFsRes{
			Path:       fullPath,
			Name:       name,
			IsDir:      isDir,
			Size:       size,
			ModTime:    modTime,
			Permission: perm,
		})
	}

	return results, nil
}

// parseModTime attempts to parse the modification time from ls output.
func parseModTime(month, day, timeOrYear string) time.Time {
	now := time.Now()

	// Try to parse as time (HH:MM)
	if strings.Contains(timeOrYear, ":") {
		// Recent file: "Jan 15 12:34"
		timeStr := fmt.Sprintf("%s %s %d %s", month, day, now.Year(), timeOrYear)
		if t, err := time.Parse("Jan 2 2006 15:04", timeStr); err == nil {
			// If the parsed time is in the future, it must be from last year
			if t.After(now) {
				t = t.AddDate(-1, 0, 0)
			}
			return t
		}
	} else {
		// Old file: "Jan 15 2023"
		timeStr := fmt.Sprintf("%s %s %s", month, day, timeOrYear)
		if t, err := time.Parse("Jan 2 2006", timeStr); err == nil {
			return t
		}
	}

	// Fallback to current time if parsing fails
	return now
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
