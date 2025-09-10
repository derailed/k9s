package model

import (
	"context"
	"log/slog"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const pulseRate = 10 * time.Second

type HealthPoint struct {
	GVR           *client.GVR
	Total, Faults int
}

type GVRs []*client.GVR

var PulseGVRs = client.GVRs{
	client.NodeGVR,
	client.NsGVR,
	client.SvcGVR,
	client.EvGVR,

	client.PodGVR,
	client.DpGVR,
	client.StsGVR,
	client.DsGVR,

	client.JobGVR,
	client.CjGVR,
	client.PvGVR,
	client.PvcGVR,

	client.HpaGVR,
	client.IngGVR,
	client.NpGVR,
	client.SaGVR,
}

func (g GVRs) First() *client.GVR {
	return g[0]
}

func (g GVRs) Last() *client.GVR {
	return g[len(g)-1]
}

func (g GVRs) Index(gvr *client.GVR) int {
	for i := range g {
		if g[i] == gvr {
			return i
		}
	}

	return -1
}

// PulseHealth tracks resources health.
type PulseHealth struct {
	factory dao.Factory
}

// NewPulseHealth returns a new instance.
func NewPulseHealth(f dao.Factory) *PulseHealth {
	return &PulseHealth{factory: f}
}

func (h *PulseHealth) Watch(ctx context.Context, ns string) HealthChan {
	c := make(HealthChan, 2)
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, false)

	go func(ctx context.Context, ns string, c HealthChan) {
		if err := h.checkPulse(ctx, ns, c); err != nil {
			slog.Error("Pulse check failed", slogs.Error, err)
		}
		for {
			select {
			case <-ctx.Done():
				close(c)
				return
			case <-time.After(pulseRate):
				if err := h.checkPulse(ctx, ns, c); err != nil {
					slog.Error("Pulse check failed", slogs.Error, err)
				}
			}
		}
	}(ctx, ns, c)

	return c
}

func (h *PulseHealth) checkPulse(ctx context.Context, ns string, c HealthChan) error {
	slog.Debug("Checking pulses...")
	for _, gvr := range PulseGVRs {
		check, err := h.check(ctx, ns, gvr)
		if err != nil {
			return err
		}
		c <- check
	}
	return nil
}

func (h *PulseHealth) check(ctx context.Context, ns string, gvr *client.GVR) (HealthPoint, error) {
	meta, ok := Registry[gvr]
	if !ok {
		meta = ResourceMeta{
			DAO:      new(dao.Table),
			Renderer: new(render.Table),
		}
	}
	if meta.DAO == nil {
		meta.DAO = &dao.Resource{}
	}

	meta.DAO.Init(h.factory, gvr)
	oo, err := meta.DAO.List(ctx, ns)
	if err != nil {
		return HealthPoint{}, err
	}
	c := HealthPoint{GVR: gvr, Total: len(oo)}
	if isTable(oo) {
		ta := oo[0].(*metav1.Table)
		c.Total = len(ta.Rows)
		for _, row := range ta.Rows {
			if err := meta.Renderer.Healthy(ctx, row); err != nil {
				c.Faults++
			}
		}
	} else {
		for _, o := range oo {
			if err := meta.Renderer.Healthy(ctx, o); err != nil {
				c.Faults++
			}
		}
	}
	slog.Debug("Checked", slogs.GVR, gvr, slogs.Config, c)

	return c, nil
}

func isTable(oo []runtime.Object) bool {
	if len(oo) == 0 || len(oo) > 1 {
		return false
	}
	_, ok := oo[0].(*metav1.Table)

	return ok
}
