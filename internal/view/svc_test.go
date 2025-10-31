// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	dao.MetaAccess.RegisterMeta(client.DirGVR.String(), &metav1.APIResource{
		Name:         "dirs",
		SingularName: "dir",
		Kind:         "Directory",
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.PodGVR.String(), &metav1.APIResource{
		Name:         "pods",
		SingularName: "pod",
		Namespaced:   true,
		Kind:         "Pods",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.NsGVR.String(), &metav1.APIResource{
		Name:         "namespaces",
		SingularName: "namespace",
		Namespaced:   true,
		Kind:         "Namespaces",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.SvcGVR.String(), &metav1.APIResource{
		Name:         "services",
		SingularName: "service",
		Namespaced:   true,
		Kind:         "Services",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.SecGVR.String(), &metav1.APIResource{
		Name:         "secrets",
		SingularName: "secret",
		Namespaced:   true,
		Kind:         "Secrets",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.PcGVR.String(), &metav1.APIResource{
		Name:         "priorityclasses",
		SingularName: "priorityclass",
		Namespaced:   false,
		Kind:         "PriorityClass",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.CmGVR.String(), &metav1.APIResource{
		Name:         "configmaps",
		SingularName: "configmap",
		Namespaced:   true,
		Kind:         "ConfigMaps",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})

	dao.MetaAccess.RegisterMeta(client.RefGVR.String(), &metav1.APIResource{
		Name:         "references",
		SingularName: "reference",
		Namespaced:   true,
		Kind:         "References",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.AliGVR.String(), &metav1.APIResource{
		Name:         "aliases",
		SingularName: "alias",
		Namespaced:   true,
		Kind:         "Aliases",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.CoGVR.String(), &metav1.APIResource{
		Name:         "containers",
		SingularName: "container",
		Namespaced:   true,
		Kind:         "Containers",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.CtGVR.String(), &metav1.APIResource{
		Name:         "contexts",
		SingularName: "context",
		Namespaced:   true,
		Kind:         "Contexts",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta("subjects", &metav1.APIResource{
		Name:         "subjects",
		SingularName: "subject",
		Namespaced:   true,
		Kind:         "Subjects",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.RbacGVR.String(), &metav1.APIResource{
		Name:         "rbacs",
		SingularName: "rbac",
		Namespaced:   true,
		Kind:         "Rbac",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.PfGVR.String(), &metav1.APIResource{
		Name:         "portforwards",
		SingularName: "portforward",
		Namespaced:   true,
		Kind:         "PortForwards",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})

	dao.MetaAccess.RegisterMeta(client.SdGVR.String(), &metav1.APIResource{
		Name:         "screendumps",
		SingularName: "screendump",
		Namespaced:   true,
		Kind:         "ScreenDumps",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.StsGVR.String(), &metav1.APIResource{
		Name:         "statefulsets",
		SingularName: "statefulset",
		Namespaced:   true,
		Kind:         "StatefulSets",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.DsGVR.String(), &metav1.APIResource{
		Name:         "daemonsets",
		SingularName: "daemonset",
		Namespaced:   true,
		Kind:         "DaemonSets",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.DpGVR.String(), &metav1.APIResource{
		Name:         "deployments",
		SingularName: "deployment",
		Namespaced:   true,
		Kind:         "Deployments",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
	dao.MetaAccess.RegisterMeta(client.PvcGVR.String(), &metav1.APIResource{
		Name:         "persistentvolumeclaims",
		SingularName: "persistentvolumeclaim",
		Namespaced:   true,
		Kind:         "PersistentVolumeClaims",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
}

func TestServiceNew(t *testing.T) {
	s := view.NewService(client.SvcGVR)

	require.NoError(t, s.Init(makeCtx(t)))
	assert.Equal(t, "Services", s.Name())
	assert.Len(t, s.Hints(), 13)
}
