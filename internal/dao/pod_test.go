package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDefaultLogContainer(t *testing.T) {
	type args struct {
		imageSpecs ImageSpecs
	}
	uu := map[string]struct {
		po   v1.Pod
		want string
	}{
		"no_annotation": {
			po:   v1.Pod{},
			want: "",
		},
		"container_not_present": {
			po:   v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"kubectl.kubernetes.io/default-logs-container": "container1"}}},
			want: "",
		},
		"container_found": {
			po:   v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"kubectl.kubernetes.io/default-logs-container": "container1"}}, Spec: v1.PodSpec{Containers: []v1.Container{{Name: "container1"}}}},
			want: "container1",
		},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			container := getDefaultLogContainer(u.po)
			assert.Equal(t, u.want, container)
		})
	}
}
