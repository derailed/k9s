package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/health"
)

// PulseListener represents a health model listener.
type PulseListener interface {
	// PulseChanged notifies the model data changed.
	PulseChanged(*health.Check)

	// PulseFailed notifies the health check failed.
	PulseFailed(error)

	// MetricsChanged update metrics time series.
	MetricsChanged(dao.TimeSeries)
}

// Pulse tracks multiple resources health.
type Pulse struct {
	gvr       *client.GVR
	namespace string
	listeners []PulseListener
	health    *PulseHealth
}

// NewPulse returns a new pulse.
func NewPulse(gvr *client.GVR) *Pulse {
	return &Pulse{
		gvr: gvr,
	}
}

type HealthChan chan HealthPoint

// Watch monitors pulses.
func (p *Pulse) Watch(ctx context.Context) (HealthChan, dao.MetricsChan, error) {
	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, nil, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	if p.health == nil {
		p.health = NewPulseHealth(f)
	}

	healthChan := p.health.Watch(ctx, p.namespace)
	metricsChan := dao.DialRecorder(f.Client()).Watch(ctx, p.namespace)

	return healthChan, metricsChan, nil
}

// Refresh update the model now.
func (*Pulse) Refresh(context.Context) {}

// GetNamespace returns the model namespace.
func (p *Pulse) GetNamespace() string {
	return p.namespace
}

// SetNamespace sets up model namespace.
func (p *Pulse) SetNamespace(ns string) {
	if client.IsAllNamespaces(ns) {
		ns = client.BlankNamespace
	}
	p.namespace = ns
}

// AddListener adds a listener.
func (p *Pulse) AddListener(l PulseListener) {
	p.listeners = append(p.listeners, l)
}

// RemoveListener delete a listener.
func (p *Pulse) RemoveListener(l PulseListener) {
	victim := -1
	for i, lis := range p.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		p.listeners = append(p.listeners[:victim], p.listeners[victim+1:]...)
	}
}
