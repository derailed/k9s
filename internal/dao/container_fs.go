// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
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
// Returns true to allow navigation; errors will be handled in List().
func (c *ContainerFs) IsDirectory(path string) bool {
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

// execInContainer executes a command in a container and returns the output using the Kubernetes API.
func (c *ContainerFs) execInContainer(ctx context.Context, podPath, container, dir string) ([]render.ContainerFsRes, error) {
	// Parse namespace and pod name
	ns, po := client.Namespaced(podPath)

	// Get REST config
	cfg, err := c.Client().RestConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	// Set up for core v1 API
	cfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	cfg.APIPath = "/api"
	cfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	// Create REST client
	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w", err)
	}

	// Build exec request
	req := restClient.Post().
		Resource("pods").
		Name(po).
		Namespace(ns).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   []string{"ls", "-la", dir},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	// Create executor
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Execute command and capture output
	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	if err != nil {
		return nil, fmt.Errorf("exec failed: %w (stderr: %s)", err, stderr.String())
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

// DownloadFile downloads a file from the container to a local path using cat.
func (c *ContainerFs) DownloadFile(ctx context.Context, podPath, container, remotePath, localPath string) error {
	// Parse namespace and pod name
	ns, po := client.Namespaced(podPath)

	// Get REST config
	cfg, err := c.Client().RestConfig()
	if err != nil {
		return fmt.Errorf("failed to get REST config: %w", err)
	}

	// Set up for core v1 API
	cfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	cfg.APIPath = "/api"
	cfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	// Create REST client
	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Build exec request to cat the file
	req := restClient.Post().
		Resource("pods").
		Name(po).
		Namespace(ns).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   []string{"cat", remotePath},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	// Create executor
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Execute cat command and write directly to file
	var stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: file,
		Stderr: &stderr,
		Tty:    false,
	})

	if err != nil {
		return fmt.Errorf("download failed: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}
