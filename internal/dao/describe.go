// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	"k8s.io/kubectl/pkg/describe"
)

// Describe describes a resource.
func Describe(c client.Connection, gvr *client.GVR, path string) (string, error) {
	mapper := RestMapper{Connection: c}
	m, err := mapper.ToRESTMapper()
	if err != nil {
		slog.Error("No REST mapper for resource",
			slogs.GVR, gvr,
			slogs.Error, err,
		)
		return "", err
	}

	gvk, err := m.KindFor(gvr.GVR())
	if err != nil {
		slog.Error("No GVK for resource %s",
			slogs.GVR, gvr,
			slogs.Error, err,
		)
		return "", err
	}

	ns, n := client.Namespaced(path)
	if client.IsClusterScoped(ns) {
		ns = client.BlankNamespace
	}
	mapping, err := mapper.ResourceFor(gvr.AsResourceName(), gvk.Kind)
	if err != nil {
		slog.Error("Unable to find mapper",
			slogs.GVR, gvr,
			slogs.ResName, n,
			slogs.Error, err,
		)
		return "", err
	}
	d, err := describe.Describer(c.Config().Flags(), mapping)
	if err != nil {
		slog.Error("Unable to find describer",
			slogs.GVR, gvr.AsResourceName(),
			slogs.Error, err,
		)
		return "", err
	}

	return d.Describe(ns, n, describe.DescriberSettings{ShowEvents: true})
}
