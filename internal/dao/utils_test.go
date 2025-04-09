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
	inventory map[string]map[*client.GVR][]runtime.Object
}

func makeFactory() dao.Factory {
	return &testFactory{
		inventory: map[string]map[*client.GVR][]runtime.Object{
			"kube-system": {
				client.SecGVR: {
					load("secret"),
				},
			},
		},
	}
}

var _ dao.Factory = &testFactory{}

func (*testFactory) Client() client.Connection {
	return nil
}
func (f *testFactory) Get(gvr *client.GVR, fqn string, _ bool, _ labels.Selector) (runtime.Object, error) {
	ns, po := path.Split(fqn)
	ns = strings.Trim(ns, "/")

	for _, o := range f.inventory[ns][gvr] {
		if o.(*unstructured.Unstructured).GetName() == po {
			return o, nil
		}
	}

	return nil, nil
}
func (f *testFactory) List(gvr *client.GVR, ns string, _ bool, _ labels.Selector) ([]runtime.Object, error) {
	return f.inventory[ns][gvr], nil
}

func (*testFactory) ForResource(string, *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}
func (*testFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (*testFactory) WaitForCacheSync() {}
func (*testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (*testFactory) DeleteForwarder(string) {}

func load(n string) *unstructured.Unstructured {
	raw, _ := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))

	var o unstructured.Unstructured
	_ = json.Unmarshal(raw, &o)

	return &o
}
