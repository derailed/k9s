// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					Annotations: map[string]string{"kubectl.kubernetes.io/default-container": "container1"},
				},
			},
			wantContainer: "",
			wantOk:        false,
		},
		"container_found": {
			po: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"kubectl.kubernetes.io/default-container": "container1"},
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
			container, ok := GetDefaultContainer(u.po.ObjectMeta, u.po.Spec)
			assert.Equal(t, u.wantContainer, container)
			assert.Equal(t, u.wantOk, ok)
		})
	}
}
