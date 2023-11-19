// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/health"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultRefreshRate = 5 * time.Second

// PulseListener represents a health model listener.
type PulseListener interface {
	// PulseChanged notifies the model data changed.
	PulseChanged(*health.Check)

	// TreeFailed notifies the health check failed.
	PulseFailed(error)
}

// Pulse tracks multiple resources health.
type Pulse struct {
	gvr         string
	namespace   string
	inUpdate    int32
	listeners   []PulseListener
	refreshRate time.Duration
	health      *PulseHealth
	data        health.Checks
}

// NewPulse returns a new pulse.
func NewPulse(gvr string) *Pulse {
	return &Pulse{
		gvr:         gvr,
		refreshRate: defaultRefreshRate,
	}
}

// Watch monitors pulses.
func (p *Pulse) Watch(ctx context.Context) {
	p.Refresh(ctx)
	go p.updater(ctx)
}

func (p *Pulse) updater(ctx context.Context) {
	defer log.Debug().Msgf("Pulse canceled -- %q", p.gvr)

	rate := initRefreshRate
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(rate):
			rate = p.refreshRate
			p.refresh(ctx)
		}
	}
}

// Refresh update the model now.
func (p *Pulse) Refresh(ctx context.Context) {
	for _, d := range p.data {
		p.firePulseChanged(d)
	}
	p.refresh(ctx)
}

func (p *Pulse) refresh(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&p.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return
	}
	defer atomic.StoreInt32(&p.inUpdate, 0)

	if err := p.reconcile(ctx); err != nil {
		log.Error().Err(err).Msg("Reconcile failed")
		p.firePulseFailed(err)
		return
	}
}

func (p *Pulse) list(ctx context.Context) ([]runtime.Object, error) {
	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	if p.health == nil {
		p.health = NewPulseHealth(f)
	}
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, false)
	return p.health.List(ctx, p.namespace)
}

func (p *Pulse) reconcile(ctx context.Context) error {
	oo, err := p.list(ctx)
	if err != nil {
		return err
	}

	p.data = health.Checks{}
	for _, o := range oo {
		c, ok := o.(*health.Check)
		if !ok {
			return fmt.Errorf("Expecting health check but got %T", o)
		}
		p.data = append(p.data, c)
		p.firePulseChanged(c)
	}
	return nil
}

// GetNamespace returns the model namespace.
func (p *Pulse) GetNamespace() string {
	return p.namespace
}

// SetNamespace sets up model namespace.
func (p *Pulse) SetNamespace(ns string) {
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

func (p *Pulse) firePulseChanged(check *health.Check) {
	for _, l := range p.listeners {
		l.PulseChanged(check)
	}
}

func (p *Pulse) firePulseFailed(err error) {
	for _, l := range p.listeners {
		l.PulseFailed(err)
	}
}
