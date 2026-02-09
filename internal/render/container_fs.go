// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ContainerFs renders a container filesystem entry to screen.
type ContainerFs struct{}

// IsGeneric identifies a generic handler.
func (ContainerFs) IsGeneric() bool {
	return false
}

// Healthy checks if the resource is healthy.
func (ContainerFs) Healthy(context.Context, any) error {
	return nil
}

// ColorerFunc colors a resource row.
func (ContainerFs) ColorerFunc() model1.ColorerFunc {
	return func(_ string, h model1.Header, re *model1.RowEvent) tcell.Color {
		// Find the NAME column to check if it's a directory
		idx, ok := h.IndexOf("NAME", true)
		if !ok {
			return model1.DefaultColorer("", h, re)
		}

		// Directories shown in blue
		if len(re.Row.Fields) > idx && len(re.Row.Fields[idx]) > 0 && re.Row.Fields[idx][0:4] == "📁 " {
			return tcell.ColorDodgerBlue
		}

		return model1.DefaultColorer("", h, re)
	}
}

func (ContainerFs) SetViewSetting(*config.ViewSetting) {}

// Header returns a header row.
func (ContainerFs) Header(string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "SIZE"},
		model1.HeaderColumn{Name: "MODIFIED"},
		model1.HeaderColumn{Name: "PERMISSIONS"},
	}
}

// Render renders a container filesystem entry to screen.
func (ContainerFs) Render(o any, _ string, r *model1.Row) error {
	cfs, ok := o.(ContainerFsRes)
	if !ok {
		return fmt.Errorf("expected ContainerFsRes, but got %T", o)
	}

	// NAME column with icon
	icon := "📄 "
	if cfs.IsDir {
		icon = "📁 "
	}
	name := icon + cfs.Name

	// SIZE column
	size := "-"
	if !cfs.IsDir {
		size = formatSize(cfs.Size)
	}

	// MODIFIED column
	modified := formatModTime(cfs.ModTime)

	// PERMISSIONS column
	permissions := cfs.Permission

	r.ID = cfs.Path
	r.Fields = append(r.Fields, name, size, modified, permissions)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// ContainerFsRes represents a container filesystem resource.
type ContainerFsRes struct {
	Path       string
	Name       string
	IsDir      bool
	Size       int64
	ModTime    time.Time
	Permission string
}

// GetObjectKind returns a schema object.
func (ContainerFsRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a copy.
func (c ContainerFsRes) DeepCopyObject() runtime.Object {
	return c
}

// formatSize formats file size in human-readable format.
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatModTime formats modification time (relative for recent, absolute for old).
func formatModTime(t time.Time) string {
	age := time.Since(t)
	if age < 24*time.Hour {
		// Relative: "2m ago", "5h ago"
		if age < time.Hour {
			mins := int(age.Minutes())
			if mins < 1 {
				return "now"
			}
			return fmt.Sprintf("%dm ago", mins)
		}
		return fmt.Sprintf("%dh ago", int(age.Hours()))
	}
	// Absolute: "Jan 15 2026"
	return t.Format("Jan 02 2006")
}
