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
	Path             string
	Container        string
	DefaultContainer string
	SinceTime        string
	Lines            int64
	SinceSeconds     int64
	Previous         bool
	SingleContainer  bool
	MultiPods        bool
	ShowTimestamp    bool
	AllContainers    bool
}

// Info returns the option pod and container info.
func (o *LogOptions) Info() string {
	return fmt.Sprintf("%q::%q", o.Path, o.Container)
}

func (o *LogOptions) Clone() *LogOptions {
	return &LogOptions{
		Path:             o.Path,
		Container:        o.Container,
		DefaultContainer: o.DefaultContainer,
		Lines:            o.Lines,
		Previous:         o.Previous,
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

// FixedSizeName returns a normalize fixed size pod name if possible.
func (o *LogOptions) FixedSizeName() string {
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
func (o *LogOptions) DecorateLog(bytes []byte) *LogItem {
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
