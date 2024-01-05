// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"regexp"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sahilm/fuzzy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetaFQN returns a fully qualified resource name.
func MetaFQN(m metav1.ObjectMeta) string {
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
func NewExpBackOff(ctx context.Context, start, max time.Duration) backoff.BackOffContext {
	bf := backoff.NewExponentialBackOff()
	bf.InitialInterval, bf.MaxElapsedTime = start, max
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
