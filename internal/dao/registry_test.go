// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMetaFor(t *testing.T) {
	uu := map[string]struct {
		gvr *client.GVR
		err error
		e   metav1.APIResource
	}{
		"xray-gvr": {
			gvr: client.XGVR,
			e: metav1.APIResource{
				Name:         "xrays",
				Kind:         "XRays",
				SingularName: "xray",
				Categories:   []string{k9sCat},
			},
		},

		"xray": {
			gvr: client.NewGVR("xrays"),
			e: metav1.APIResource{
				Name:         "xrays",
				Kind:         "XRays",
				SingularName: "xray",
				Categories:   []string{k9sCat},
			},
		},

		"policy": {
			gvr: client.NewGVR("policy"),
			e: metav1.APIResource{
				Name:       "policies",
				Kind:       "Rules",
				Namespaced: true,
				Categories: []string{k9sCat},
			},
		},

		"helm": {
			gvr: client.NewGVR("helm"),
			e: metav1.APIResource{
				Name:       "helm",
				Kind:       "Helm",
				Namespaced: true,
				Verbs:      []string{"delete"},
				Categories: []string{helmCat},
			},
		},

		"toast": {
			gvr: client.NewGVR("blah"),
			err: errors.New("no resource meta defined for\n \"blah\""),
		},
	}

	m := NewMeta()
	require.NoError(t, m.LoadResources(nil))
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			meta, err := m.MetaFor(u.gvr)
			assert.Equal(t, u.err, err)
			if err == nil {
				assert.Equal(t, &u.e, meta)
			}
		})
	}
}

func TestAddCRDPropertiesMarksK8sIOCRD(t *testing.T) {
	gatewayGVR := client.NewGVR("gateway.networking.k8s.io/v1/httproutes")
	gatewayMeta := &metav1.APIResource{
		Name:       "httproutes",
		Group:      "gateway.networking.k8s.io",
		Version:    "v1",
		Kind:       "HTTPRoute",
		Categories: []string{"gateway-api"},
	}
	ingressGVR := client.NewGVR("networking.k8s.io/v1/ingresses")
	ingressMeta := &metav1.APIResource{
		Name:    "ingresses",
		Group:   "networking.k8s.io",
		Version: "v1",
		Kind:    "Ingress",
	}
	metas := ResourceMetas{
		gatewayGVR: gatewayMeta,
		ingressGVR: ingressMeta,
	}
	crd := apiext.CustomResourceDefinition{
		Spec: apiext.CustomResourceDefinitionSpec{
			Group: "gateway.networking.k8s.io",
			Names: apiext.CustomResourceDefinitionNames{
				Plural: "httproutes",
				Kind:   "HTTPRoute",
			},
			Versions: []apiext.CustomResourceDefinitionVersion{
				{Name: "v1", Served: true},
			},
		},
	}

	addCRDProperties(metas, &crd)

	assert.True(t, IsCRD(gatewayMeta))
	assert.Equal(t, []string{"gateway-api", crdCat}, gatewayMeta.Categories)
	assert.False(t, IsCRD(ingressMeta))
}
