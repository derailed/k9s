// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
)

// RefScanner represents a resource reference scanner.
type RefScanner interface {
	// Init initializes the scanner
	Init(Factory, *client.GVR)

	// Scan scan the resource for references.
	Scan(ctx context.Context, gvr *client.GVR, fqn string, wait bool) (Refs, error)

	// ScanSA scan the resource for serviceaccount references.
	ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error)
}

// Ref represents a resource reference.
type Ref struct {
	GVR string
	FQN string
}

// Refs represents a collection of resource references.
type Refs []Ref

var (
	_ RefScanner = (*Deployment)(nil)
	_ RefScanner = (*StatefulSet)(nil)
	_ RefScanner = (*DaemonSet)(nil)
	_ RefScanner = (*Job)(nil)
	_ RefScanner = (*CronJob)(nil)
)

func scanners() map[*client.GVR]RefScanner {
	return map[*client.GVR]RefScanner{
		client.DpGVR:  new(Deployment),
		client.DsGVR:  new(DaemonSet),
		client.StsGVR: new(StatefulSet),
		client.CjGVR:  new(CronJob),
		client.JobGVR: new(Job),
	}
}

// ScanForRefs scans cluster resources for resource references.
func ScanForRefs(ctx context.Context, f Factory) (Refs, error) {
	rgvr, ok := ctx.Value(internal.KeyGVR).(*client.GVR)
	if !ok {
		return nil, errors.New("expecting context GVR")
	}
	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("expecting context Path")
	}
	wait, ok := ctx.Value(internal.KeyWait).(bool)
	if !ok {
		slog.Warn("Expecting context Wait key. Using default")
	}

	var wg sync.WaitGroup
	out := make(chan Refs)
	for gvr, scanner := range scanners() {
		wg.Add(1)
		go func(ctx context.Context, gvr *client.GVR, s RefScanner, out chan Refs, wait bool) {
			defer wg.Done()
			s.Init(f, gvr)
			refs, err := s.Scan(ctx, rgvr, fqn, wait)
			if err != nil {
				slog.Error("Reference scan failed for",
					slogs.RefType, fmt.Sprintf("%T", s),
					slogs.Error, err,
				)
				return
			}
			select {
			case out <- refs:
			case <-ctx.Done():
				return
			}
		}(ctx, gvr, scanner, out, wait)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	res := make(Refs, 0, 10)
	for refs := range out {
		res = append(res, refs...)
	}

	return res, nil
}

// ScanForSARefs scans cluster resources for serviceaccount refs.
func ScanForSARefs(ctx context.Context, f Factory) (Refs, error) {
	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("expecting context Path")
	}
	wait, ok := ctx.Value(internal.KeyWait).(bool)
	if !ok {
		return nil, errors.New("expecting context Wait")
	}

	var wg sync.WaitGroup
	out := make(chan Refs)
	for gvr, scanner := range scanners() {
		wg.Add(1)
		go func(ctx context.Context, gvr *client.GVR, s RefScanner, out chan Refs, wait bool) {
			defer wg.Done()
			s.Init(f, gvr)
			refs, err := s.ScanSA(ctx, fqn, wait)
			if err != nil {
				slog.Error("ServiceAccount scan failed",
					slogs.RefType, fmt.Sprintf("%T", s),
					slogs.Error, err,
				)
				return
			}
			select {
			case out <- refs:
			case <-ctx.Done():
				return
			}
		}(ctx, gvr, scanner, out, wait)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	res := make(Refs, 0, 10)
	for refs := range out {
		res = append(res, refs...)
	}

	return res, nil
}
