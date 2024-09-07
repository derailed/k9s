// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	appsv1 "k8s.io/api/apps/v1"
)

// StatefulSet represents a statefulset viewer.
type StatefulSet struct {
	ResourceViewer
}

// NewStatefulSet returns a new viewer.
func NewStatefulSet(gvr client.GVR) ResourceViewer {
	var s StatefulSet
	s.ResourceViewer = NewPortForwardExtender(
		NewVulnerabilityExtender(
			NewRestartExtender(
				NewScaleExtender(
					NewImageExtender(
						NewOwnerExtender(
							NewLogsExtender(NewBrowser(gvr), s.logOptions),
						),
					),
				),
			),
		),
	)
	s.AddBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showPods)

	return &s
}

func (s *StatefulSet) logOptions(prev bool) (*dao.LogOptions, error) {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return nil, errors.New("you must provide a selection")
	}
	sts, err := s.getInstance(path)
	if err != nil {
		return nil, err
	}

	return podLogOptions(s.App(), path, prev, sts.ObjectMeta, sts.Spec.Template.Spec), nil
}

func (s *StatefulSet) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyShiftR, ui.NewKeyAction("Sort Ready", s.GetTable().SortColCmd(readyCol, true), false))
}

func (s *StatefulSet) showPods(app *App, _ ui.Tabular, _ client.GVR, path string) {
	i, err := s.getInstance(path)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, path, i.Spec.Selector)
}

func (s *StatefulSet) getInstance(path string) (*appsv1.StatefulSet, error) {
	var sts dao.StatefulSet

	return sts.GetInstance(s.App().factory, path)
}
