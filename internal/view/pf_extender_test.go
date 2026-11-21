// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func TestEnsurePodPortFwdAllowed(t *testing.T) {
	uu := map[string]struct {
		podExists   bool
		podPhase    corev1.PodPhase
		expectError bool
	}{
		"pod-not-exist": {
			expectError: true,
		},
		"pod-pending": {
			podExists:   true,
			podPhase:    corev1.PodPending,
			expectError: true,
		},
		"pod-running": {
			podExists:   true,
			podPhase:    corev1.PodRunning,
			expectError: false,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			f := testFactory{}
			if u.podExists {
				f.expectedGet = &unstructured.Unstructured{
					Object: map[string]any{
						"status": map[string]any{
							"phase": u.podPhase,
						},
					},
				}
			}

			err := ensurePodPortFwdAllowed(f, "ns/name")
			if u.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPortForwardExtenderShowPFFromService(t *testing.T) {
	const (
		svcPath = "default/my-svc"
		podPath = "default/web-7d8b49557c-x2v5k"
		pfID    = podPath + "|web|8080:80"
	)

	app := NewApp(mock.NewMockConfig(t))
	app.BenchFile = "bench.yaml"
	app.factory = newPFTestFactory(t,
		&unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata":   map[string]any{"namespace": "default", "name": "my-svc"},
			"spec":       map[string]any{"selector": map[string]any{"app": "web"}},
		}},
		&unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]any{
				"namespace": "default",
				"name":      "web-7d8b49557c-x2v5k",
				"labels":    map[string]any{"app": "web"},
			},
		}},
	)
	fwd := dao.NewPortForwarder(app.factory)
	app.factory.Forwarders()[pfID] = fwd
	app.factory.Forwarders()["default/other-0|other|9090:90"] = dao.NewPortForwarder(app.factory)

	p := PortForwardExtender{ResourceViewer: &pfViewerStub{
		app:   app,
		gvr:   client.SvcGVR,
		table: newPFTestTable(t, client.SvcGVR, svcPath),
	}}
	require.Equal(t, svcPath, p.GetTable().GetSelectedItem())

	assert.Nil(t, p.showPFCmd(nil))
	top, ok := app.Content.Top().(*PortForward)
	require.True(t, ok, "expected the port-forward view to be pushed")
	b, ok := top.ResourceViewer.(*Browser)
	require.True(t, ok)
	require.NotNil(t, b.contextFn)

	var pf dao.PortForward
	pf.Init(app.factory, client.PfGVR)
	oo, err := pf.List(b.contextFn(context.Background()), "")
	require.NoError(t, err)
	require.Len(t, oo, 1, "the service pod's port-forward should be listed")
	assert.Same(t, fwd, oo[0].(render.ForwardRes).Forwarder)
}

type pfViewerStub struct {
	ResourceViewer

	app   *App
	gvr   *client.GVR
	table *Table
}

func (s *pfViewerStub) App() *App        { return s.app }
func (s *pfViewerStub) GVR() *client.GVR { return s.gvr }
func (s *pfViewerStub) GetTable() *Table { return s.table }

func newPFTestTable(t *testing.T, gvr *client.GVR, path string) *Table {
	tv := NewTable(gvr)
	require.NoError(t, tv.Init(makeContext(t)))
	tv.SetModel(&mockTableModel{})
	data := model1.NewTableDataWithRows(
		gvr,
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME"},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{Row: model1.Row{ID: path, Fields: model1.Fields{"default", "my-svc"}}},
		),
	)
	tv.UpdateUI(tv.Update(data, false), data)
	tv.Select(1, 0)

	return tv
}

type pfTestConn struct {
	client.Connection

	dial dynamic.Interface
}

func (pfTestConn) CanI(string, *client.GVR, string, []string) (bool, error) { return true, nil }
func (c pfTestConn) DynDial() (dynamic.Interface, error)                    { return c.dial, nil }

func newPFTestFactory(t *testing.T, oo ...runtime.Object) *watch.Factory {
	dial := dynfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			client.SvcGVR.GVR(): "ServiceList",
			client.PodGVR.GVR(): "PodList",
		},
		oo...,
	)
	f := watch.NewFactory(pfTestConn{dial: dial})
	f.Start("default")
	t.Cleanup(f.Terminate)
	for _, gvr := range []*client.GVR{client.SvcGVR, client.PodGVR} {
		inf, err := f.ForResource("default", gvr)
		require.NoError(t, err)
		require.True(t, cache.WaitForCacheSync(t.Context().Done(), inf.Informer().HasSynced))
	}

	return f
}

type testFactory struct {
	expectedGet runtime.Object
}

var _ dao.Factory = testFactory{}

func (testFactory) Client() client.Connection {
	return nil
}
func (t testFactory) Get(*client.GVR, string, bool, labels.Selector) (runtime.Object, error) {
	if t.expectedGet != nil {
		return t.expectedGet, nil
	}

	return nil, errors.New("not found")
}
func (testFactory) List(*client.GVR, string, bool, labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (testFactory) ForResource(string, *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}
func (testFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (testFactory) WaitForCacheSync()      {}
func (testFactory) DeleteForwarder(string) {}
