// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port

import (
	"errors"
)

type Annotations map[string]string

func (a Annotations) PreferredPorts(specs ContainerPortSpecs) (PFAnns, error) {
	if len(specs) == 0 {
		return nil, errors.New("no exposed ports")
	}

	value, ok := a[K9sPortForwardsKey]
	if !ok {
		return PFAnns{specs[0].ToPFAnn()}, nil
	}

	return specs.MatchAnnotations(value), nil
}
