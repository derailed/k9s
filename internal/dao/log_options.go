package dao

import (
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogOptions represent logger options.
type LogOptions struct {
	Path            string
	Container       string
	Lines           int64
	Previous        bool
	SingleContainer bool
	MultiPods       bool
	ShowTimestamp   bool
	SinceTime       string
	SinceSeconds    int64
	In, Out         string
}

// HasContainer checks if a container is present.
func (o LogOptions) HasContainer() bool {
	return o.Container != ""
}

// ToPodLogOptions returns pod log options.
func (o LogOptions) ToPodLogOptions() *v1.PodLogOptions {
	opts := v1.PodLogOptions{
		Follow:     true,
		Timestamps: true,
		Container:  o.Container,
		Previous:   o.Previous,
		TailLines:  &o.Lines,
	}

	if o.SinceSeconds < 0 {
		return &opts
	}
	if o.SinceSeconds != 0 {
		opts.SinceSeconds = &o.SinceSeconds
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

// FixedSizeName returns a normalize fixed size pod name if possible.
func (o LogOptions) FixedSizeName() string {
	_, n := client.Namespaced(o.Path)
	tokens := strings.Split(n, "-")
	if len(tokens) < 3 {
		return n
	}
	var s []string
	for i := 0; i < len(tokens)-1; i++ {
		s = append(s, tokens[i])
	}

	return Truncate(strings.Join(s, "-"), 15) + "-" + tokens[len(tokens)-1]
}

// DecorateLog add a log header to display po/co information along with the log message.
func (o LogOptions) DecorateLog(bytes []byte) *LogItem {
	item := NewLogItem(bytes)
	if len(bytes) == 0 {
		return item
	}

	if o.MultiPods {
		_, pod := client.Namespaced(o.Path)
		item.Pod, item.Container = pod, o.Container
	}

	if !o.SingleContainer {
		item.Container = o.Container
	}

	return item
}
