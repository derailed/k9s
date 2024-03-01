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
						NewLogsExtender(NewBrowser(gvr), s.logOptions),
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

	cc := sts.Spec.Template.Spec.Containers
	var (
		co, dco string
		allCos  bool
	)
	if c, ok := dao.GetDefaultContainer(sts.Spec.Template.ObjectMeta, sts.Spec.Template.Spec); ok {
		co, dco = c, c
	} else if len(cc) == 1 {
		co = cc[0].Name
	} else {
		dco, allCos = cc[0].Name, true
	}

	cfg := s.App().Config.K9s.Logger
	opts := dao.LogOptions{
		Path:            path,
		Container:       co,
		Lines:           int64(cfg.TailCount),
		SingleContainer: len(cc) == 1,
		SinceSeconds:    cfg.SinceSeconds,
		AllContainers:   allCos,
		ShowTimestamp:   cfg.ShowTime,
		Previous:        prev,
	}
	if co == "" {
		opts.AllContainers = true
	}
	opts.DefaultContainer = dco

	return &opts, nil
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
