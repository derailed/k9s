// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package watch

import (
	"fmt"
	"log/slog"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func toGVR(gvr string) schema.GroupVersionResource {
	tokens := strings.Split(gvr, "/")
	if len(tokens) < 3 {
		tokens = append([]string{""}, tokens...)
	}

	return schema.GroupVersionResource{
		Group:    tokens[0],
		Version:  tokens[1],
		Resource: tokens[2],
	}
}

func namespaced(n string) (ns, res string) {
	ns, res = path.Split(n)

	return strings.Trim(ns, "/"), res
}

// DumpFactory for debug.
func DumpFactory(f *Factory) {
	slog.Debug("----------- FACTORIES -------------")
	for ns := range f.factories {
		slog.Debug(fmt.Sprintf("  Factory for NS %q", ns))
	}
	slog.Debug("-----------------------------------")
}

// DebugFactory for debug.
func DebugFactory(f *Factory, ns, gvr string) {
	slog.Debug(fmt.Sprintf("----------- DEBUG FACTORY (%s) -------------", gvr))
	fac, ok := f.factories[ns]
	if !ok {
		return
	}
	inf := fac.ForResource(toGVR(gvr))
	for i, k := range inf.Informer().GetStore().ListKeys() {
		slog.Debug(fmt.Sprintf("%d -- %s", i, k))
	}
}
