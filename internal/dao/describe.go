// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"k8s.io/kubectl/pkg/describe"
)

// Describe describes a resource.
func Describe(c client.Connection, gvr client.GVR, path string) (string, error) {
	mapper := RestMapper{Connection: c}
	m, err := mapper.ToRESTMapper()
	if err != nil {
		log.Error().Err(err).Msgf("No REST mapper for resource %s", gvr)
		return "", err
	}

	gvk, err := m.KindFor(gvr.GVR())
	if err != nil {
		log.Error().Err(err).Msgf("No GVK for resource %s", gvr)
		return "", err
	}

	ns, n := client.Namespaced(path)
	if client.IsClusterScoped(ns) {
		ns = client.BlankNamespace
	}
	mapping, err := mapper.ResourceFor(gvr.AsResourceName(), gvk.Kind)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find mapper for %s %s", gvr, n)
		return "", err
	}
	d, err := describe.Describer(c.Config().Flags(), mapping)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find describer for %#v", mapping)
		return "", err
	}

	return d.Describe(ns, n, describe.DescriberSettings{ShowEvents: true})
}
