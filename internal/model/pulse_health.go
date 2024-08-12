// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/health"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PulseHealth tracks resources health.
type PulseHealth struct {
	factory dao.Factory
}

// NewPulseHealth returns a new instance.
func NewPulseHealth(f dao.Factory) *PulseHealth {
	return &PulseHealth{
		factory: f,
	}
}

// List returns a canned collection of resources health.
func (h *PulseHealth) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	gvrs := []string{
		"v1/pods",
		"v1/events",
		"apps/v1/replicasets",
		"apps/v1/deployments",
		"apps/v1/statefulsets",
		"apps/v1/daemonsets",
		"batch/v1/jobs",
		"v1/persistentvolumes",
	}

	hh := make([]runtime.Object, 0, 10)
	for _, gvr := range gvrs {
		c, err := h.check(ctx, ns, gvr)
		if err != nil {
			return nil, err
		}
		hh = append(hh, c)
	}

	mm, err := h.checkMetrics(ctx)
	if err != nil {
		return hh, err
	}
	for _, m := range mm {
		hh = append(hh, m)
	}

	return hh, nil
}

func (h *PulseHealth) checkMetrics(ctx context.Context) (health.Checks, error) {
	dial := client.DialMetrics(h.factory.Client())

	nn, err := dao.FetchNodes(ctx, h.factory, "")
	if err != nil {
		return nil, err
	}

	nmx, err := dial.FetchNodesMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching metrics")
		return nil, err
	}

	mx := make(client.NodesMetrics, len(nn.Items))
	dial.NodesMetrics(nn, nmx, mx)

	var ccpu, cmem, acpu, amem, tcpu, tmem int64
	for _, m := range mx {
		ccpu += m.CurrentCPU
		cmem += m.CurrentMEM
		acpu += m.AllocatableCPU
		amem += m.AllocatableMEM
		tcpu += m.TotalCPU
		tmem += m.TotalMEM
	}
	c1 := health.NewCheck("cpu")
	c1.Set(health.S1, ccpu)
	c1.Set(health.S2, acpu)
	c1.Set(health.S3, tcpu)
	c2 := health.NewCheck("mem")
	c2.Set(health.S1, cmem)
	c2.Set(health.S2, amem)
	c2.Set(health.S3, tmem)

	return health.Checks{c1, c2}, nil
}

func (h *PulseHealth) check(ctx context.Context, ns, gvr string) (*health.Check, error) {
	meta, ok := Registry[gvr]
	if !ok {
		meta = ResourceMeta{
			DAO:      &dao.Table{},
			Renderer: &render.Generic{},
		}
	}
	if meta.DAO == nil {
		meta.DAO = &dao.Resource{}
	}

	meta.DAO.Init(h.factory, client.NewGVR(gvr))
	oo, err := meta.DAO.List(ctx, ns)
	if err != nil {
		return nil, err
	}
	c := health.NewCheck(gvr)

	if meta.Renderer.IsGeneric() {
		table, ok := oo[0].(*metav1.Table)
		if !ok {
			return nil, fmt.Errorf("expecting a meta table but got %T", oo[0])
		}
		rows := make(model1.Rows, len(table.Rows))
		re, _ := meta.Renderer.(model1.Generic)
		re.SetTable(ns, table)
		for i, row := range table.Rows {
			if err := re.Render(row, ns, &rows[i]); err != nil {
				return nil, err
			}
			if !model1.IsValid(ns, re.Header(ns), rows[i]) {
				c.Inc(health.S2)
				continue
			}
			c.Inc(health.S1)
		}
		c.Total(int64(len(table.Rows)))
		return c, nil
	}
	c.Total(int64(len(oo)))
	rr, re := make(model1.Rows, len(oo)), meta.Renderer
	for i, o := range oo {
		if err := re.Render(o, ns, &rr[i]); err != nil {
			return nil, err
		}
		if !model1.IsValid(ns, re.Header(ns), rr[i]) {
			c.Inc(health.S2)
			continue
		}
		c.Inc(health.S1)
	}

	return c, nil
}
