// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/sahilm/fuzzy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getMeta(ctx context.Context, gvr *client.GVR) (ResourceMeta, error) {
	meta := resourceMeta(gvr)
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return ResourceMeta{}, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	meta.DAO.Init(factory, gvr)

	return meta, nil
}

func resourceMeta(gvr *client.GVR) ResourceMeta {
	meta, ok := Registry[gvr]
	if !ok {
		meta = ResourceMeta{
			DAO:      new(dao.Table),
			Renderer: new(render.Table),
		}
	}
	if meta.DAO == nil {
		meta.DAO = new(dao.Resource)
	}

	return meta
}

// MetaFQN returns a fully qualified resource name.
func MetaFQN(m *metav1.ObjectMeta) string {
	return FQN(m.Namespace, m.Name)
}

// FQN returns a fully qualified resource name.
func FQN(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

// NewExpBackOff returns a new exponential backoff timer.
func NewExpBackOff(ctx context.Context, start, maxVal time.Duration) backoff.BackOffContext {
	bf := backoff.NewExponentialBackOff()
	bf.InitialInterval, bf.MaxElapsedTime = start, maxVal
	return backoff.WithContext(bf, ctx)
}

func rxFilter(q string, lines []string) fuzzy.Matches {
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return nil
	}

	matches := make(fuzzy.Matches, 0, len(lines))
	for i, l := range lines {
		locs := rx.FindAllStringIndex(l, -1)
		for _, loc := range locs {
			indexes := make([]int, 0, loc[1]-loc[0])
			for v := loc[0]; v < loc[1]; v++ {
				indexes = append(indexes, v)
			}

			matches = append(matches, fuzzy.Match{Str: q, Index: i, MatchedIndexes: indexes})
		}
	}

	return matches
}
