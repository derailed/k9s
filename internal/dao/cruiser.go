// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func mustMap(o runtime.Object, field string) map[string]interface{} {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		panic("no unstructured")
	}
	m, ok := u.Object[field].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("map extract failed for %q", field))
	}

	return m
}

func mustSlice(o runtime.Object, field string) []interface{} {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return []interface{}{}
	}
	s, ok := u.Object[field].([]interface{})
	if !ok {
		return []interface{}{}
	}

	return s
}

func mustField(o map[string]interface{}, field string) interface{} {
	f, ok := o[field]
	if !ok {
		panic(fmt.Sprintf("no field for %q", field))
	}

	return f
}
