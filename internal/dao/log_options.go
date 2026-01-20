// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogOptions represents logger options.
type LogOptions struct {
	CreateDuration   time.Duration
	Path             string
	Container        string
	DefaultContainer string
	SinceTime        string
	Lines            int64
	SinceSeconds     int64
	Head             bool
	Previous         bool
	SingleContainer  bool
	MultiPods        bool
	ShowTimestamp    bool
	AllContainers    bool
}

// Info returns the option pod and container info.
func (o *LogOptions) Info() string {
	if o.Container != "" {
		return fmt.Sprintf("%s (%s)", o.Path, o.Container)
	}

	return o.Path
}

// Clone clones options.
func (o *LogOptions) Clone() *LogOptions {
	return &LogOptions{
		Path:             o.Path,
		Container:        o.Container,
		DefaultContainer: o.DefaultContainer,
		Lines:            o.Lines,
		Previous:         o.Previous,
		Head:             o.Head,
		SingleContainer:  o.SingleContainer,
		MultiPods:        o.MultiPods,
		ShowTimestamp:    o.ShowTimestamp,
		SinceTime:        o.SinceTime,
		SinceSeconds:     o.SinceSeconds,
		AllContainers:    o.AllContainers,
	}
}

// HasContainer checks if a container is present.
func (o *LogOptions) HasContainer() bool {
	return o.Container != ""
}

// ToggleAllContainers toggles single or all-containers if possible.
func (o *LogOptions) ToggleAllContainers() {
	if o.SingleContainer {
		return
	}
	o.AllContainers = !o.AllContainers
	if o.AllContainers {
		o.DefaultContainer, o.Container = o.Container, ""
		return
	}

	if o.DefaultContainer != "" {
		o.Container = o.DefaultContainer
	}
}

// ToPodLogOptions returns pod log options.
func (o *LogOptions) ToPodLogOptions() *v1.PodLogOptions {
	opts := v1.PodLogOptions{
		Follow:     true,
		Timestamps: true,
		Container:  o.Container,
		Previous:   o.Previous,
		TailLines:  &o.Lines,
	}
	if o.Head {
		var maxBytes int64 = 5000
		opts.Follow = false
		opts.TailLines, opts.SinceSeconds, opts.SinceTime = nil, nil, nil
		opts.LimitBytes = &maxBytes
		return &opts
	}
	if o.SinceSeconds < 0 {
		return &opts
	}

	if o.SinceSeconds != 0 {
		opts.SinceSeconds, opts.SinceTime = &o.SinceSeconds, nil
		return &opts
	}

	if o.SinceTime == "" {
		return &opts
	}
	if t, err := time.Parse(time.RFC3339, o.SinceTime); err == nil {
		opts.SinceTime = &metav1.Time{Time: t.Add(time.Second)}
	}

	return &opts
}

// ToLogItem add a log header to display po/co information along with the log message.
func (o *LogOptions) ToLogItem(bytes []byte) *LogItem {
	item := NewLogItem(bytes)
	if len(bytes) == 0 {
		return item
	}
	item.SingleContainer = o.SingleContainer
	if item.SingleContainer {
		item.Container = o.Container
	}
	if o.MultiPods {
		_, pod := client.Namespaced(o.Path)
		item.Pod, item.Container = pod, o.Container
	} else {
		item.Container = o.Container
	}

	return item
}

func (*LogOptions) ToErrLogItem(err error) *LogItem {
	t := time.Now().UTC().Format(time.RFC3339Nano)
	item := NewLogItem([]byte(fmt.Sprintf("%s [orange::b]%s[::-]\n", t, err)))
	item.IsError = true
	return item
}

// ToKubectlCommand generates the equivalent kubectl command for these log options.
func (o *LogOptions) ToKubectlCommand() string {
	ns, pod := client.Namespaced(o.Path)

	var parts []string
	parts = append(parts, "kubectl", "logs")

	// Add namespace if present
	if ns != "" {
		parts = append(parts, "-n", ns)
	}

	// Add pod name
	parts = append(parts, pod)

	// Add container flag
	if o.AllContainers {
		parts = append(parts, "--all-containers=true")
	} else if o.Container != "" {
		parts = append(parts, "-c", o.Container)
	}

	// Add previous flag
	if o.Previous {
		parts = append(parts, "--previous")
	}

	// Handle head mode (limit bytes)
	if o.Head {
		parts = append(parts, "--limit-bytes=5000")
	} else {
		// Add tail lines
		if o.Lines > 0 {
			parts = append(parts, fmt.Sprintf("--tail=%d", o.Lines))
		}

		// Add since time or since seconds
		if o.SinceSeconds > 0 {
			// Convert seconds to human-readable format
			if o.SinceSeconds < 60 {
				parts = append(parts, fmt.Sprintf("--since=%ds", o.SinceSeconds))
			} else if o.SinceSeconds < 3600 {
				parts = append(parts, fmt.Sprintf("--since=%dm", o.SinceSeconds/60))
			} else {
				parts = append(parts, fmt.Sprintf("--since=%dh", o.SinceSeconds/3600))
			}
		} else if o.SinceTime != "" {
			parts = append(parts, fmt.Sprintf("--since-time=%s", o.SinceTime))
		}

		// Add follow flag (k9s always follows)
		parts = append(parts, "-f")
	}

	// Add timestamps flag
	if o.ShowTimestamp {
		parts = append(parts, "--timestamps")
	}

	return strings.Join(parts, " ")
}
