package model

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/health"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
)

type Health struct {
	factory dao.Factory
}

func NewHealth(f dao.Factory) *Health {
	return &Health{
		factory: f,
	}
}

func (h *Health) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("HealthCheck %v", time.Since(t))
	}(time.Now())

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

	mm, err := h.checkMetrics()
	if err != nil {
		return hh, nil
	}
	for _, m := range mm {
		hh = append(hh, m)
	}

	return hh, nil
}

func (h *Health) checkMetrics() (health.Checks, error) {
	dial := client.DialMetrics(h.factory.Client())
	nmx, err := dial.FetchNodesMetrics()
	if err != nil {
		log.Error().Err(err).Msgf("Fetching metrics")
		return nil, err
	}

	var cpu, mem float64
	for _, mx := range nmx.Items {
		cpu += float64(mx.Usage.Cpu().MilliValue())
		mem += client.ToMB(mx.Usage.Memory().Value())
	}
	c1 := health.NewCheck("cpu")
	c1.Set(health.OK, int(math.Round(cpu)))
	c2 := health.NewCheck("mem")
	c2.Set(health.OK, int(math.Round(mem)))

	return health.Checks{c1, c2}, nil
}

func (h *Health) check(ctx context.Context, ns, gvr string) (*health.Check, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("  CHECK %s - %v", gvr, time.Since(t))
	}(time.Now())

	meta, ok := Registry[gvr]
	if !ok {
		return nil, fmt.Errorf("No meta for %q", gvr)
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
	c.Total(len(oo))
	rr, re := make(render.Rows, len(oo)), meta.Renderer
	for i, o := range oo {
		if err := re.Render(o, ns, &rr[i]); err != nil {
			return nil, err
		}
		if !render.Happy(ns, rr[i]) {
			c.Inc(health.Toast)
		} else {
			c.Inc(health.OK)
		}
	}

	return c, nil
}
