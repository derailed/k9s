package dao

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/describe/versioned"
)

// Describe describes a resource.
func Describe(c client.Connection, gvr client.GVR, path string) (string, error) {
	log.Debug().Msgf("DESCRIBE %q::%q", gvr, path)
	mapper := RestMapper{Connection: c}
	m, err := mapper.ToRESTMapper()
	if err != nil {
		log.Error().Err(err).Msgf("No REST mapper for resource %s", gvr)
		return "", err
	}

	gvk, err := m.KindFor(gvr.AsGVR())
	if err != nil {
		log.Error().Err(err).Msgf("No GVK for resource %s", gvr)
		return "", err
	}

	ns, n := client.Namespaced(path)
	if client.IsClusterScoped(ns) {
		ns = client.AllNamespaces
	}
	mapping, err := mapper.ResourceFor(gvr.AsResourceName(), gvk.Kind)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find mapper for %s %s", gvr, n)
		return "", err
	}
	d, err := versioned.Describer(c.Config().Flags(), mapping)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find describer for %#v", mapping)
		return "", err
	}

	return d.Describe(ns, n, describe.DescriberSettings{ShowEvents: true})
}
