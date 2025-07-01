// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func mustMap(o runtime.Object, field string) map[string]any {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		panic("no unstructured")
	}
	m, ok := u.Object[field].(map[string]any)
	if !ok {
		panic(fmt.Sprintf("map extract failed for %q", field))
	}

	return m
}

func mustSlice(o runtime.Object, field string) []any {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}
	s, ok := u.Object[field].([]any)
	if !ok {
		return nil
	}

	return s
}

func mustField(o map[string]any, field string) any {
	f, ok := o[field]
	if !ok {
		panic(fmt.Sprintf("no field for %q", field))
	}

	return f
}
