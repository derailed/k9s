// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// DeltaSign signals a diff.
	DeltaSign = "Δ"
	// PlusSign signals inc.
	PlusSign = "[red::b]↑"
	// MinusSign signal dec.
	MinusSign = "[green::b]↓"
)

var percent = regexp.MustCompile(`\A(\d+)\%\z`)

func deltaNumb(o, n string) (string, bool) {
	var delta string

	i, ok := numerical(o)
	if !ok {
		return delta, ok
	}

	j, _ := numerical(n)
	switch {
	case i < j:
		delta = PlusSign
	case i > j:
		delta = MinusSign
	}

	return delta, ok
}

func deltaPerc(o, n string) (string, bool) {
	var delta string
	i, ok := percentage(o)
	if !ok {
		return delta, ok
	}

	j, _ := percentage(n)
	switch {
	case i < j:
		delta = PlusSign
	case i > j:
		delta = MinusSign
	}

	return delta, ok
}

func deltaQty(o, n string) (string, bool) {
	var delta string
	q1, err := resource.ParseQuantity(o)
	if err != nil {
		return delta, false
	}

	q2, _ := resource.ParseQuantity(n)
	switch q1.Cmp(q2) {
	case -1:
		delta = PlusSign
	case 1:
		delta = MinusSign
	}
	return delta, true
}

func deltaDur(o, n string) (string, bool) {
	var delta string
	d1, err := time.ParseDuration(o)
	if err != nil {
		return delta, false
	}

	d2, _ := time.ParseDuration(n)
	switch {
	case d2-d1 > 0:
		delta = PlusSign
	case d2-d1 < 0:
		delta = MinusSign
	}
	return delta, true
}

// Deltas signals diffs between 2 strings.
func Deltas(o, n string) string {
	o, n = strings.TrimSpace(o), strings.TrimSpace(n)
	if o == "" || o == render.NAValue {
		return ""
	}

	if d, ok := deltaNumb(o, n); ok {
		return d
	}

	if d, ok := deltaPerc(o, n); ok {
		return d
	}

	if d, ok := deltaQty(o, n); ok {
		return d
	}

	if d, ok := deltaDur(o, n); ok {
		return d
	}

	switch strings.Compare(o, n) {
	case 1, -1:
		return DeltaSign
	default:
		return ""
	}
}

func percentage(s string) (int, bool) {
	if res := percent.FindStringSubmatch(s); len(res) == 2 {
		n, _ := strconv.Atoi(res[1])
		return n, true
	}

	return 0, false
}

func numerical(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	return n, true
}
