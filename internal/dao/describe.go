package dao

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/describe/versioned"
)

func Describe(c k8s.Connection, gvr GVR, ns, n string) (string, error) {
	mapper := k8s.RestMapper{Connection: c}
	m, err := mapper.ToRESTMapper()
	if err != nil {
		log.Error().Err(err).Msgf("No REST mapper for resource %s", gvr)
		return "", err
	}

	GVR := k8s.GVR(gvr)
	gvk, err := m.KindFor(GVR.AsGVR())
	if err != nil {
		log.Error().Err(err).Msgf("No GVK for resource %s", gvr)
		return "", err
	}

	mapping, err := mapper.ResourceFor(GVR.ResName(), gvk.Kind)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find mapper for %s %s", gvr, n)
		return "", err
	}
	d, err := versioned.Describer(c.Config().Flags(), mapping)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to find describer for %#v", mapping)
		return "", err
	}

	log.Debug().Msgf("DESCRIBE FOR %q -- %q:%q", gvr, ns, n)
	return d.Describe(ns, n, describe.DescriberSettings{ShowEvents: true})
}
