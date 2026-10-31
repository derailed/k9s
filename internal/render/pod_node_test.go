// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func pinnedAffinity(key string, op v1.NodeSelectorOperator, values ...string) *v1.Affinity {
	return &v1.Affinity{NodeAffinity: &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{{
				MatchFields: []v1.NodeSelectorRequirement{{Key: key, Operator: op, Values: values}},
			}},
		},
	}}
}

func TestPodNodeName(t *testing.T) {
	uu := map[string]struct {
		spec v1.PodSpec
		e    string
	}{
		"bound": {
			spec: v1.PodSpec{NodeName: "n1", Affinity: pinnedAffinity("metadata.name", v1.NodeSelectorOpIn, "n2")},
			e:    "n1",
		},
		"pending-daemonset": {
			spec: v1.PodSpec{Affinity: pinnedAffinity("metadata.name", v1.NodeSelectorOpIn, "n2")},
			e:    "n2",
		},
		"pending-no-affinity": {
			spec: v1.PodSpec{},
		},
		"pending-other-field": {
			spec: v1.PodSpec{Affinity: pinnedAffinity("metadata.labels", v1.NodeSelectorOpIn, "n2")},
		},
		"pending-multiple-values": {
			spec: v1.PodSpec{Affinity: pinnedAffinity("metadata.name", v1.NodeSelectorOpIn, "n2", "n3")},
		},
		"pending-not-in": {
			spec: v1.PodSpec{Affinity: pinnedAffinity("metadata.name", v1.NodeSelectorOpNotIn, "n2")},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, render.PodNodeName(&u.spec))
		})
	}
}
