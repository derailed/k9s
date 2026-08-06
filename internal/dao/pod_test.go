// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
)

func TestGetDefaultContainer(t *testing.T) {
	uu := map[string]struct {
		po            *v1.Pod
		wantContainer string
		wantOk        bool
	}{
		"no_annotation": {
			po: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "container1"}},
				},
			},
			wantContainer: "",
			wantOk:        false,
		},
		"container_not_present": {
			po: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{DefaultContainerAnnotation: "container1"},
				},
			},
			wantContainer: "",
			wantOk:        false,
		},
		"container_found": {
			po: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{DefaultContainerAnnotation: "container1"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "container1"}},
				},
			},
			wantContainer: "container1",
			wantOk:        true,
		},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			container, ok := GetDefaultContainer(&u.po.ObjectMeta, &u.po.Spec)
			assert.Equal(t, u.wantContainer, container)
			assert.Equal(t, u.wantOk, ok)
		})
	}
}

func TestMatchesFieldSelector(t *testing.T) {
	pod := &unstructured.Unstructured{
		Object: map[string]any{
			"metadata": map[string]any{
				"name":      "runner-abc123",
				"namespace": "gha-runners",
			},
			"spec": map[string]any{
				"nodeName": "node-1",
			},
			"status": map[string]any{
				"phase": "Running",
			},
		},
	}

	uu := map[string]struct {
		selector string
		want     bool
	}{
		"empty_selector_matches_all": {
			selector: "",
			want:     true,
		},
		"metadata_name_match": {
			selector: "metadata.name=runner-abc123",
			want:     true,
		},
		"metadata_name_mismatch": {
			selector: "metadata.name=some-other-pod",
			want:     false,
		},
		"spec_nodeName_match": {
			selector: "spec.nodeName=node-1",
			want:     true,
		},
		"spec_nodeName_mismatch": {
			selector: "spec.nodeName=node-2",
			want:     false,
		},
		"status_phase_not_equal": {
			selector: "status.phase!=Pending",
			want:     true,
		},
		"multiple_requirements_all_match": {
			selector: "metadata.namespace=gha-runners,metadata.name=runner-abc123",
			want:     true,
		},
		"multiple_requirements_one_mismatch": {
			selector: "metadata.namespace=gha-runners,metadata.name=nope",
			want:     false,
		},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			sel, err := fields.ParseSelector(u.selector)
			assert.NoError(t, err)
			assert.Equal(t, u.want, matchesFieldSelector(pod, sel))
		})
	}
}
