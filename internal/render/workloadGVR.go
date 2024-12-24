// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type WorkloadGVR struct {
	Base
}

// Header returns a header row.
func (WorkloadGVR) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NANE"},
		model1.HeaderColumn{Name: "INCONTEXT"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (wgvr WorkloadGVR) Render(o interface{}, ns string, r *model1.Row) error {
	res, ok := o.(WorkloadGVRRes)
	if !ok {
		return fmt.Errorf("expected WorkloadGVRRes, but got %T", o)
	}

	r.ID = res.Filepath.Name()
	r.Fields = model1.Fields{
		strings.TrimSuffix(res.Filepath.Name(), filepath.Ext(res.Filepath.Name())),
		strconv.FormatBool(res.InContext),
		"",
		timeToAge(res.Filepath.ModTime()),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

type WorkloadGVRRes struct {
	Filepath  os.FileInfo
	InContext bool
}

// GetObjectKind returns a schema object.
func (a WorkloadGVRRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (a WorkloadGVRRes) DeepCopyObject() runtime.Object {
	return a
}
