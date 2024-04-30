// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	appsvalpha1 "github.com/apecloud/kubeblocks/apis/apps/v1alpha1"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"k8s.io/apimachinery/pkg/labels"
)

// Component represents a statefulset viewer.
type Component struct {
	ResourceViewer
}

// NewStatefulSet returns a new viewer.
func NewComponent(gvr client.GVR) ResourceViewer {
	var s Component
	s.ResourceViewer = NewPortForwardExtender(
		NewLogsExtender(NewBrowser(gvr), nil),
	)

	s.AddBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showPods)

	return &s
}

func (s *Component) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyShiftR, ui.NewKeyAction("Sort Ready", s.GetTable().SortColCmd(readyCol, true), false))
}

func (s *Component) showPods(app *App, _ ui.Tabular, _ client.GVR, path string) {
	i, err := s.getInstance(path)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, path, labels.Set(i.Labels).AsSelector().String(), "")
}

func (s *Component) getInstance(path string) (*appsvalpha1.Component, error) {
	var sts dao.Component

	return sts.GetInstance(s.App().factory, path)
}
