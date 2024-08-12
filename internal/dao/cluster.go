// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
)

// RefScanner represents a resource reference scanner.
type RefScanner interface {
	// Init initializes the scanner
	Init(Factory, client.GVR)
	// Scan scan the resource for references.
	Scan(ctx context.Context, gvr client.GVR, fqn string, wait bool) (Refs, error)
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
	// _ RefScanner = (*Pod)(nil)
)

func scanners() map[string]RefScanner {
	return map[string]RefScanner{
		"apps/v1/deployments":  &Deployment{},
		"apps/v1/statefulsets": &StatefulSet{},
		"apps/v1/daemonsets":   &DaemonSet{},
		"batch/v1/jobs":        &Job{},
		"batch/v1/cronjobs":    &CronJob{},
		// "v1/pods":              &Pod{},
	}
}

// ScanForRefs scans cluster resources for resource references.
func ScanForRefs(ctx context.Context, f Factory) (Refs, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("Cluster Scan %v", time.Since(t))
	}(time.Now())

	gvr, ok := ctx.Value(internal.KeyGVR).(client.GVR)
	if !ok {
		return nil, errors.New("expecting context GVR")
	}
	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("expecting context Path")
	}
	wait, ok := ctx.Value(internal.KeyWait).(bool)
	if !ok {
		log.Error().Msgf("expecting Context Wait Key")
	}

	ss := scanners()
	var wg sync.WaitGroup
	wg.Add(len(ss))
	out := make(chan Refs)
	for k, s := range ss {
		go func(ctx context.Context, kind string, s RefScanner, out chan Refs, wait bool) {
			defer wg.Done()
			s.Init(f, client.NewGVR(kind))
			refs, err := s.Scan(ctx, gvr, fqn, wait)
			if err != nil {
				log.Error().Err(err).Msgf("scan failed for %T", s)
				return
			}
			select {
			case out <- refs:
			case <-ctx.Done():
				return
			}
		}(ctx, k, s, out, wait)
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
	defer func(t time.Time) {
		log.Debug().Msgf("SA Cluster Scan %v", time.Since(t))
	}(time.Now())

	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("expecting context Path")
	}
	wait, ok := ctx.Value(internal.KeyWait).(bool)
	if !ok {
		return nil, errors.New("expecting context Wait")
	}

	ss := scanners()
	var wg sync.WaitGroup
	wg.Add(len(ss))
	out := make(chan Refs)
	for k, s := range ss {
		go func(ctx context.Context, kind string, s RefScanner, out chan Refs, wait bool) {
			defer wg.Done()
			s.Init(f, client.NewGVR(kind))
			refs, err := s.ScanSA(ctx, fqn, wait)
			if err != nil {
				log.Error().Err(err).Msgf("scan failed for %T", s)
				return
			}
			select {
			case out <- refs:
			case <-ctx.Done():
				return
			}
		}(ctx, k, s, out, wait)
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
