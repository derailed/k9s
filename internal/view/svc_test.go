// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	dao.MetaAccess.RegisterMeta("dir", metav1.APIResource{
		Name:         "dir",
		SingularName: "dir",
		Kind:         "Directory",
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("v1/pods", metav1.APIResource{
		Name:         "pods",
		SingularName: "pod",
		Namespaced:   true,
		Kind:         "Pods",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("v1/namespaces", metav1.APIResource{
		Name:         "namespaces",
		SingularName: "namespace",
		Namespaced:   true,
		Kind:         "Namespaces",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("v1/services", metav1.APIResource{
		Name:         "services",
		SingularName: "service",
		Namespaced:   true,
		Kind:         "Services",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("v1/secrets", metav1.APIResource{
		Name:         "secrets",
		SingularName: "secret",
		Namespaced:   true,
		Kind:         "Secrets",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("scheduling.k8s.io/v1/priorityclasses", metav1.APIResource{
		Name:         "priorityclasses",
		SingularName: "priorityclass",
		Namespaced:   false,
		Kind:         "PriorityClass",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("v1/configmaps", metav1.APIResource{
		Name:         "configmaps",
		SingularName: "configmap",
		Namespaced:   true,
		Kind:         "ConfigMaps",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})

	dao.MetaAccess.RegisterMeta("references", metav1.APIResource{
		Name:         "references",
		SingularName: "reference",
		Namespaced:   true,
		Kind:         "References",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("aliases", metav1.APIResource{
		Name:         "aliases",
		SingularName: "alias",
		Namespaced:   true,
		Kind:         "Aliases",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("containers", metav1.APIResource{
		Name:         "containers",
		SingularName: "container",
		Namespaced:   true,
		Kind:         "Containers",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("contexts", metav1.APIResource{
		Name:         "contexts",
		SingularName: "context",
		Namespaced:   true,
		Kind:         "Contexts",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("subjects", metav1.APIResource{
		Name:         "subjects",
		SingularName: "subject",
		Namespaced:   true,
		Kind:         "Subjects",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("rbac", metav1.APIResource{
		Name:         "rbacs",
		SingularName: "rbac",
		Namespaced:   true,
		Kind:         "Rbac",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("portforwards", metav1.APIResource{
		Name:         "portforwards",
		SingularName: "portforward",
		Namespaced:   true,
		Kind:         "PortForwards",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})

	dao.MetaAccess.RegisterMeta("screendumps", metav1.APIResource{
		Name:         "screendumps",
		SingularName: "screendump",
		Namespaced:   true,
		Kind:         "ScreenDumps",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("apps/v1/statefulsets", metav1.APIResource{
		Name:         "statefulsets",
		SingularName: "statefulset",
		Namespaced:   true,
		Kind:         "StatefulSets",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("apps/v1/daemonsets", metav1.APIResource{
		Name:         "daemonsets",
		SingularName: "daemonset",
		Namespaced:   true,
		Kind:         "DaemonSets",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("apps/v1/deployments", metav1.APIResource{
		Name:         "deployments",
		SingularName: "deployment",
		Namespaced:   true,
		Kind:         "Deployments",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("v1/persistentvolumeclaims", metav1.APIResource{
		Name:         "persistentvolumeclaims",
		SingularName: "persistentvolumeclaim",
		Namespaced:   true,
		Kind:         "PersistentVolumeClaims",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
}

func TestServiceNew(t *testing.T) {
	s := view.NewService(client.NewGVR("v1/services"))

	assert.Nil(t, s.Init(makeCtx()))
	assert.Equal(t, "Services", s.Name())
	assert.Equal(t, 12, len(s.Hints()))
}
