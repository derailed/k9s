// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/fvbommel/sortorder"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
)

// parallelRender fans work across NumCPU batch workers.
func parallelRender(n int, fn func(i int) error) error {
	if n == 0 {
		return nil
	}

	workers := min(runtime.NumCPU(), n)
	if workers < 1 {
		workers = 1
	}
	chunkSize := (n + workers - 1) / workers

	var (
		wg       sync.WaitGroup
		firstErr error
		errOnce  sync.Once
	)
	for w := range workers {
		lo := w * chunkSize
		hi := min(lo+chunkSize, n)
		if lo >= n {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				if err := fn(i); err != nil {
					errOnce.Do(func() { firstErr = err })
				}
			}
		}()
	}
	wg.Wait()

	return firstErr
}

func Hydrate(ns string, oo []k8sruntime.Object, rr Rows, re Renderer) error {
	return parallelRender(len(oo), func(i int) error {
		return re.Render(oo[i], ns, &rr[i])
	})
}

func GenericHydrate(ns string, table *metav1.Table, rr Rows, re Renderer) error {
	gr, ok := re.(Generic)
	if !ok {
		return fmt.Errorf("expecting generic renderer but got %T", re)
	}
	gr.SetTable(ns, table)

	return parallelRender(len(table.Rows), func(i int) error {
		return gr.Render(table.Rows[i], ns, &rr[i])
	})
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
	keys = make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals = make([]string, 0, len(m))
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
	d1, d2 := durationToSeconds(s1), durationToSeconds(s2)
	return d1 <= d2
}

func lessCapacity(s1, s2 string) bool {
	c1, c2 := capacityToNumber(s1), capacityToNumber(s2)

	return c1 <= c2
}

func lessNumber(s1, s2 string) bool {
	v1, v2 := strings.ReplaceAll(s1, ",", ""), strings.ReplaceAll(s2, ",", "")

	return sortorder.NaturalLess(v1, v2)
}
