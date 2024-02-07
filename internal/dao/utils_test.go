// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/watch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

type testFactory struct {
	inventory map[string]map[string][]runtime.Object
}

func makeFactory() dao.Factory {
	return &testFactory{
		inventory: map[string]map[string][]runtime.Object{
			"kube-system": {
				"v1/secrets": {
					load("secret"),
				},
			},
		},
	}
}

var _ dao.Factory = &testFactory{}

func (f *testFactory) Client() client.Connection {
	return nil
}
func (f *testFactory) Get(gvr, fqn string, wait bool, sel labels.Selector) (runtime.Object, error) {
	ns, po := path.Split(fqn)
	ns = strings.Trim(ns, "/")

	for _, o := range f.inventory[ns][gvr] {
		if o.(*unstructured.Unstructured).GetName() == po {
			return o, nil
		}
	}

	return nil, nil
}
func (f *testFactory) List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error) {
	return f.inventory[ns][gvr], nil
}

func (f *testFactory) ForResource(ns, gvr string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f *testFactory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f *testFactory) WaitForCacheSync() {}
func (f *testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (f *testFactory) DeleteForwarder(string) {}

func load(n string) *unstructured.Unstructured {
	raw, _ := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))

	var o unstructured.Unstructured
	_ = json.Unmarshal(raw, &o)

	return &o
}
