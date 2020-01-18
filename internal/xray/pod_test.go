package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodRender(t *testing.T) {
	uu := map[string]struct {
		file           string
		level1, level2 int
		status         string
	}{
		"plain": {
			file:   "po",
			level1: 1,
			level2: 1,
			status: xray.OkStatus,
		},
		"withInit": {
			file:   "init",
			level1: 1,
			level2: 1,
			status: xray.OkStatus,
		},
	}

	var re xray.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := load(t, u.file)
			root := xray.NewTreeNode("pods", "pods")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			assert.Nil(t, re.Render(ctx, "", &render.PodWithMetrics{Raw: o}))
			assert.Equal(t, u.level1, root.Size())
			assert.Equal(t, u.level2, root.Children[0].Size())
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makePod(n string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}
}

func makePodEnv(n, ref string, optional bool) v1.Pod {
	po := makePod(n)
	po.Spec.Containers = []v1.Container{
		{
			Name: "c1",
			Env: []v1.EnvVar{
				{
					Name: "e1",
					ValueFrom: &v1.EnvVarSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm1",
							},
							Key:      "k1",
							Optional: &optional,
						},
					},
				},
			},
		},
		{
			Name: "c2",
			Env: []v1.EnvVar{
				{
					Name: "e2",
					ValueFrom: &v1.EnvVarSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm2",
							},
							Key:      "k2",
							Optional: &optional,
						},
					},
				},
			},
		},
	}
	po.Spec.InitContainers = []v1.Container{
		{
			Name: "ic1",
			Env: []v1.EnvVar{
				{
					Name: "e1",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: "sec2"},
							Key:                  "k2",
							Optional:             &optional,
						},
					},
				},
			},
		},
	}

	return po
}

func makePodStatus(n, ref string, optional bool) v1.Pod {
	po := makePod(n)
	po.Status = v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionTrue,
			},
		},
		ContainerStatuses: []v1.ContainerStatus{
			{
				Name:  "c1",
				State: v1.ContainerState{Running: &v1.ContainerStateRunning{}},
			},
		},
	}
	po.Spec.Containers = []v1.Container{
		{
			Name: "c1",
			Env: []v1.EnvVar{
				{
					Name: "e1",
					ValueFrom: &v1.EnvVarSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm1",
							},
							Key:      "k1",
							Optional: &optional,
						},
					},
				},
			},
		},
		{
			Name: "c2",
			Env: []v1.EnvVar{
				{
					Name: "e2",
					ValueFrom: &v1.EnvVarSource{
						ConfigMapKeyRef: &v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm2",
							},
							Key:      "k2",
							Optional: &optional,
						},
					},
				},
			},
		},
	}
	po.Spec.InitContainers = []v1.Container{
		{
			Name: "ic1",
			Env: []v1.EnvVar{
				{
					Name: "e1",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: "sec2"},
							Key:                  "k2",
							Optional:             &optional,
						},
					},
				},
			},
		},
	}

	return po
}
