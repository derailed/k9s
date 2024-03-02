// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/fvbommel/sortorder"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Hydrate(ns string, oo []runtime.Object, rr Rows, re Renderer) error {
	for i, o := range oo {
		if err := re.Render(o, ns, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

func GenericHydrate(ns string, table *metav1.Table, rr Rows, re Renderer) error {
	gr, ok := re.(Generic)
	if !ok {
		return fmt.Errorf("expecting generic renderer but got %T", re)
	}
	gr.SetTable(ns, table)
	for i, row := range table.Rows {
		if err := gr.Render(row, ns, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

// IsValid returns true if resource is valid, false otherwise.
func IsValid(ns string, h Header, r Row) bool {
	if len(r.Fields) == 0 {
		return true
	}
	idx, ok := h.IndexOf("VALID", true)
	if !ok || idx >= len(r.Fields) {
		return true
	}

	return strings.TrimSpace(r.Fields[idx]) == ""
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
	if len(duration) == 0 {
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
	v1, v2 := strings.Replace(s1, ",", "", -1), strings.Replace(s2, ",", "", -1)

	return sortorder.NaturalLess(v1, v2)
}
