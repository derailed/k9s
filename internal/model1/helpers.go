// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/fvbommel/sortorder"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const poolSize = 10

func Hydrate(ns string, oo []runtime.Object, rr Rows, re Renderer) error {
	pool := NewWorkerPool(context.Background(), poolSize)
	for i, o := range oo {
		pool.Add(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				slog.Debug("Worker canceled")
				return nil
			default:
				return re.Render(o, ns, &rr[i])
			}
		})
	}
	errs := pool.Drain()
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func GenericHydrate(ns string, table *metav1.Table, rr Rows, re Renderer) error {
	gr, ok := re.(Generic)
	if !ok {
		return fmt.Errorf("expecting generic renderer but got %T", re)
	}
	gr.SetTable(ns, table)
	pool := NewWorkerPool(context.Background(), poolSize)
	for i, row := range table.Rows {
		pool.Add(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				slog.Debug("Worker canceled")
				return nil
			default:
				return gr.Render(row, ns, &rr[i])
			}
		})
	}
	errs := pool.Drain()
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

// IsValid returns true if resource is valid, false otherwise.
func IsValid(_ string, h Header, r Row) bool {
	if len(r.Fields) == 0 {
		return true
	}
	idx, ok := h.IndexOf("VALID", true)
	if !ok || idx >= len(r.Fields) {
		return true
	}

	return strings.TrimSpace(r.Fields[idx]) == "" || strings.ToLower(strings.TrimSpace(r.Fields[idx])) == "true"
}

func sortLabels(m map[string]string) (keys, vals []string) {
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, m[k])
	}

	return
}

// Converts labels string to map.
func labelize(labels string) map[string]string {
	ll := strings.Split(labels, ",")
	data := make(map[string]string, len(ll))
	for _, l := range ll {
		tokens := strings.Split(l, "=")
		if len(tokens) == 2 {
			data[tokens[0]] = tokens[1]
		}
	}

	return data
}

func durationToSeconds(duration string) int64 {
	if duration == "" {
		return 0
	}
	if duration == NAValue {
		return math.MaxInt64
	}

	num := make([]rune, 0, 5)
	var n, m int64
	for _, r := range duration {
		switch r {
		case 'y':
			m = 365 * 24 * 60 * 60
		case 'd':
			m = 24 * 60 * 60
		case 'h':
			m = 60 * 60
		case 'm':
			m = 60
		case 's':
			m = 1
		default:
			num = append(num, r)
			continue
		}
		n, num = n+runesToNum(num)*m, num[:0]
	}

	return n
}

func runesToNum(rr []rune) int64 {
	var r int64
	var m int64 = 1
	for i := len(rr) - 1; i >= 0; i-- {
		v := int64(rr[i] - '0')
		r += v * m
		m *= 10
	}

	return r
}

func capacityToNumber(capacity string) int64 {
	if strings.TrimSpace(capacity) == "" {
		return 0
	}
	quantity := resource.MustParse(capacity)
	return quantity.Value()
}

// Less return true if c1 <= c2.
func Less(isNumber, isDuration, isCapacity bool, id1, id2, v1, v2 string) bool {
	var less bool
	switch {
	case isNumber:
		less = lessNumber(v1, v2)
	case isDuration:
		less = lessDuration(v1, v2)
	case isCapacity:
		less = lessCapacity(v1, v2)
	default:
		less = sortorder.NaturalLess(v1, v2)
	}
	if v1 == v2 {
		return sortorder.NaturalLess(id1, id2)
	}

	return less
}

func lessDuration(s1, s2 string) bool {
	// Try timestamps first for full-precision comparison.
	t1, e1 := time.Parse(time.RFC3339, s1)
	t2, e2 := time.Parse(time.RFC3339, s2)
	if e1 == nil && e2 == nil {
		return !t1.Before(t2)
	}

	// Fall back to humanized duration parsing.
	d1, d2 := durationToSeconds(s1), durationToSeconds(s2)
	return d1 <= d2
}

// lessTimestamp compares two rows by stashed timestamps for the given column name.
// Returns (less, true) if both rows have a stashed timestamp, or (false, false) to
// signal the caller to fall back to string-based comparison.
func lessTimestamp(r1, r2 Row, colName, id1, id2 string) (less, ok bool) {
	t1, ok1 := r1.Timestamps[colName]
	t2, ok2 := r2.Timestamps[colName]
	if !ok1 || !ok2 {
		return false, false
	}
	if t1.Equal(t2) {
		return sortorder.NaturalLess(id1, id2), true
	}
	// Newer timestamp means younger (smaller age).
	return !t1.Before(t2), true
}

func lessCapacity(s1, s2 string) bool {
	c1, c2 := capacityToNumber(s1), capacityToNumber(s2)

	return c1 <= c2
}

func lessNumber(s1, s2 string) bool {
	v1, v2 := strings.ReplaceAll(s1, ",", ""), strings.ReplaceAll(s2, ",", "")

	return sortorder.NaturalLess(v1, v2)
}
