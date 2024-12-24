// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*WorkloadGVR)(nil)

type WorkloadGVR struct {
	NonResource
}

func NewWorkloadGVR(f Factory) *WorkloadGVR {
	a := WorkloadGVR{}
	a.Init(f, client.NewGVR("workloadGVR"))

	return &a
}

// List returns a collection of aliases.
func (a *WorkloadGVR) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	workloadsDir, _ := ctx.Value(internal.KeyDir).(string)
	clusterContext, _ := ctx.Value(internal.KeyPath).(string)

	// List files from custom workload directory
	ff, err := os.ReadDir(workloadsDir)
	if err != nil {
		return nil, err
	}

	// Generate workload list from custom gvrs
	oo := make([]runtime.Object, len(ff))
	for i, f := range ff {
		if fi, err := f.Info(); err == nil {
			oo[i] = render.WorkloadGVRRes{
				Filepath:  fi,
				InContext: a.isInContext(clusterContext, fi.Name())}
		}
	}

	return oo, nil
}

func (a *WorkloadGVR) isInContext(ctxPath, filename string) bool {
	// Read cluster context config
	content, err := os.ReadFile(ctxPath)
	if err != nil {
		return false
	}

	// Unmarshal cluster config
	var config config.WorkloadConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return false
	}

	// Check if custom GVR is in context
	for _, n := range config.GVRFilenames {
		if n == strings.TrimSuffix(filename, filepath.Ext(filename)) {
			return true
		}
	}

	return false

}

// Get fetch a resource.
func (a *WorkloadGVR) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, errors.New("nyi")
}
