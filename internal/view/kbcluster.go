// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	appsvalpha1 "github.com/apecloud/kubeblocks/apis/apps/v1alpha1"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
)

// Cluster represents a statefulset viewer.
type Cluster struct {
	ResourceViewer
}

// NewStatefulSet returns a new viewer.
func NewCluster(gvr client.GVR) ResourceViewer {
	var s Cluster
	s.ResourceViewer = NewPortForwardExtender(
		NewLogsExtender(NewBrowser(gvr), nil),
	)

	s.AddBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showComponents)

	return &s
}

func (s *Cluster) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyShiftR, ui.NewKeyAction("Sort Ready", s.GetTable().SortColCmd(readyCol, true), false))
}

func (s *Cluster) showComponents(app *App, _ ui.Tabular, _ client.GVR, path string) {
	i, err := s.getInstance(path)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	lables := map[string]string{
		"app.kubernetes.io/instance": i.GetName(),
	}
	showComponents(app, path, lables)
}

func (s *Cluster) getInstance(path string) (*appsvalpha1.Cluster, error) {
	var sts dao.Cluster
	return sts.GetInstance(s.App().factory, path)
}

func showComponents(app *App, path string, labelSel map[string]string) {
	v := NewComponent(client.NewGVR("apps.kubeblocks.io/v1alpha1/components"))
	v.SetContextFn(podCtx(app, path, ""))
	v.SetLabelFilter(labelSel)

	ns, _ := client.Namespaced(path)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config NS set failed!")
	}
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}
