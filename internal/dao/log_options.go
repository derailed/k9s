// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	bytes "bytes"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
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
	DecodeJson       bool
	Json             JsonOptions
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
		DecodeJson:       o.DecodeJson,
		Json:             o.Json,
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

// HandleJson if JSON decoding is turned on, processes JSON templates with JQ
func (o *LogOptions) HandleJson(bb []byte) []byte {
	if !o.DecodeJson {
		return bb
	}

	var b bytes.Buffer
	orgLine := string(bb)
	result, _ := o.Json.GetCompiledJsonQuery().Run(orgLine).Next()
	err := PrintJsonResult(result, &b)
	if err != nil {
		log.Trace().AnErr("Error", err).Msg("Error printing JQ result")
		fmt.Fprintf(&b, "%s: %s\n", "JQ", err)
		fmt.Fprintf(&b, "Original line: %s", orgLine)
	}
	if b.Bytes()[b.Len()-1] != '\n' {
		b.WriteByte('\n')
	}
	return b.Bytes()
}

func PrintJsonResult(value any, outStream io.Writer) error {
	if err, ok := value.(error); ok {
		return err
	}
	if s, ok := value.(string); ok {
		_, err := outStream.Write([]byte(s))
		return err
	}
	return errors.New("JQ result not a string")
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
