// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

type workloadTestConn struct {
	client.Connection
	dyn  dynamic.Interface
	kube kubernetes.Interface
}

func (c workloadTestConn) DynDial() (dynamic.Interface, error)                    { return c.dyn, nil }
func (c workloadTestConn) Dial() (kubernetes.Interface, error)                    { return c.kube, nil }
func (workloadTestConn) CanI(string, *client.GVR, string, []string) (bool, error) { return true, nil }
func (workloadTestConn) ConnectionOK() bool                                       { return true }
func (workloadTestConn) HasMetrics() bool                                         { return false }

func toUnstructured(t *testing.T, obj any, apiVersion, kind string) *unstructured.Unstructured {
	t.Helper()
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	require.NoError(t, err)
	u := &unstructured.Unstructured{Object: m}
	u.SetAPIVersion(apiVersion)
	u.SetKind(kind)
	return u
}

func makeTestDeployment(name string, replicas *int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "app", Image: "nginx:1.25"},
					},
				},
			},
		},
	}
}

func makeTestStatefulSet(name string, replicas *int32) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "db", Image: "postgres:16"},
					},
				},
			},
		},
	}
}

func setupWorkloadView(t *testing.T, objs ...runtime.Object) (*Workload, *k8sfake.Clientset) {
	t.Helper()

	prevMeta := dao.MetaAccess
	dao.MetaAccess = dao.NewMeta()
	t.Cleanup(func() { dao.MetaAccess = prevMeta })

	dao.MetaAccess.RegisterMeta(client.PodGVR.String(), &metav1.APIResource{
		Name:       "pods",
		Kind:       "Pods",
		Namespaced: true,
		Verbs:      metav1.Verbs{"get", "list", "watch"},
	})
	dao.MetaAccess.RegisterMeta(client.RsGVR.String(), &metav1.APIResource{
		Name:       "replicasets",
		Group:      "apps",
		Version:    "v1",
		Kind:       "ReplicaSets",
		Namespaced: true,
		Verbs:      metav1.Verbs{"get", "list", "watch"},
	})
	dao.MetaAccess.RegisterMeta(client.SvcGVR.String(), &metav1.APIResource{
		Name:       "services",
		Kind:       "Services",
		Namespaced: true,
		Verbs:      metav1.Verbs{"get", "list", "watch"},
	})
	dao.MetaAccess.RegisterMeta(client.WkGVR.String(), &metav1.APIResource{
		Name:       "workloads",
		Kind:       "Workloads",
		Namespaced: true,
		Categories: []string{"k9s"},
		Verbs:      metav1.Verbs{"get", "list"},
	})

	var (
		dynObjs  []runtime.Object
		kubeObjs []runtime.Object
	)
	for _, o := range objs {
		kubeObjs = append(kubeObjs, o)
		switch v := o.(type) {
		case *appsv1.Deployment:
			dynObjs = append(dynObjs, toUnstructured(t, v, "apps/v1", "Deployment"))
		case *appsv1.StatefulSet:
			dynObjs = append(dynObjs, toUnstructured(t, v, "apps/v1", "StatefulSet"))
		case *corev1.Pod:
			dynObjs = append(dynObjs, toUnstructured(t, v, "v1", "Pod"))
		}
	}

	dynClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			client.DpGVR.GVR():  "DeploymentList",
			client.StsGVR.GVR(): "StatefulSetList",
			client.PodGVR.GVR(): "PodList",
			client.RsGVR.GVR():  "ReplicaSetList",
			client.SvcGVR.GVR(): "ServiceList",
		},
		dynObjs...,
	)
	kubeClient := k8sfake.NewClientset(kubeObjs...)

	cfg := mock.NewMockConfig(t)
	conn := workloadTestConn{
		Connection: cfg.GetConnection(),
		dyn:        dynClient,
		kube:       kubeClient,
	}
	app := NewApp(cfg)
	app.factory = watch.NewFactory(conn)
	app.factory.Start("default")
	t.Cleanup(app.factory.Terminate)
	app.command = NewCommand(app)
	app.command.alias = dao.NewAlias(app.factory)
	app.command.alias.Define(client.SvcGVR, client.SvcGVR.String(), "svc", "services")
	prevViewers := customViewers
	t.Cleanup(func() { customViewers = prevViewers })
	if customViewers == nil {
		customViewers = loadCustomViewers()
	}

	w := NewWorkload(client.WkGVR).(*Workload)
	ctx := context.WithValue(context.Background(), internal.KeyApp, app)
	ctx = context.WithValue(ctx, internal.KeyStyles, app.Styles)
	require.NoError(t, w.Init(ctx))

	return w, kubeClient
}

func selectWorkloadRow(w *Workload, rowID string) {
	data := model1.NewTableDataWithRows(
		client.WkGVR,
		model1.Header{
			model1.HeaderColumn{Name: "KIND"},
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME"},
		},
		model1.NewRowEventsWithEvts(model1.RowEvent{
			Row: model1.Row{
				ID:     rowID,
				Fields: model1.Fields{"kind", "default", "name"},
			},
		}),
	)
	w.GetTable().SetModel(&staticTableModel{data: data})
	cell := tview.NewTableCell("kind").SetReference(rowID)
	w.GetTable().SetCell(0, 0, tview.NewTableCell("KIND"))
	w.GetTable().SetCell(1, 0, cell)
	w.GetTable().Select(1, 0)
}

type staticTableModel struct {
	mockTableModel
	data *model1.TableData
}

func (m *staticTableModel) Empty() bool             { return m.data.Empty() }
func (m *staticTableModel) RowCount() int           { return m.data.RowCount() }
func (m *staticTableModel) Peek() *model1.TableData { return m.data.Clone() }

func TestWorkloadBindKeysReadOnly(t *testing.T) {
	uu := map[string]struct {
		readOnly bool
	}{
		"read-write": {readOnly: false},
		"read-only":  {readOnly: true},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cfg := mock.NewMockConfig(t)
			cfg.K9s.ReadOnly = u.readOnly
			app := NewApp(cfg)
			w := NewWorkload(client.WkGVR).(*Workload)
			ctx := context.WithValue(context.Background(), internal.KeyApp, app)
			ctx = context.WithValue(ctx, internal.KeyStyles, app.Styles)
			require.NoError(t, w.GetTable().Init(ctx))

			aa := ui.NewKeyActions()
			w.bindKeys(aa)

			for _, key := range []tcell.Key{ui.KeyL, ui.KeyP, ui.KeyF, ui.KeyShiftF, ui.KeyZ, ui.KeyY, ui.KeyD} {
				_, ok := aa.Get(key)
				assert.True(t, ok, "expected key %v to be bound", key)
			}
			for _, key := range []tcell.Key{ui.KeyE, tcell.KeyCtrlD, ui.KeyR, ui.KeyS, ui.KeyI} {
				_, ok := aa.Get(key)
				assert.Equal(t, !u.readOnly, ok, "unexpected binding state for dangerous key %v", key)
			}
		})
	}
}

func TestWorkloadUnsupportedKindActionsAreNoOp(t *testing.T) {
	w, _ := setupWorkloadView(t)
	evt := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)

	for _, id := range []string{
		"v1/pods|default|p1",
		"v1/services|default|svc1",
		"apps/v1/daemonsets|default|ds1",
		"apps/v1/replicasets|default|rs1",
	} {
		selectWorkloadRow(w, id)
		assert.Equal(t, evt, w.logsCmd(false)(evt), "logs should be no-op for %s", id)
		assert.Equal(t, evt, w.logsCmd(true)(evt), "logs prev should be no-op for %s", id)
		assert.Equal(t, evt, w.restartCmd(evt), "restart should be no-op for %s", id)
		assert.Equal(t, evt, w.scaleCmd(evt), "scale should be no-op for %s", id)
		assert.Equal(t, evt, w.setImageCmd(evt), "setImage should be no-op for %s", id)
		assert.Equal(t, evt, w.portFwdCmd(evt), "portFwd should be no-op for %s", id)
		assert.Equal(t, evt, w.showPFCmd(evt), "showPF should be no-op for %s", id)
		assert.Equal(t, evt, w.replicaSetsCmd(evt), "replicaSets should be no-op for %s", id)
	}

	// StatefulSet does not support ReplicaSets ('z').
	selectWorkloadRow(w, "apps/v1/statefulsets|default|db")
	assert.Equal(t, evt, w.replicaSetsCmd(evt))
}

func TestWorkloadShowRes(t *testing.T) {
	replicas := int32(2)
	dp := makeTestDeployment("web", &replicas)
	sts := makeTestStatefulSet("db", &replicas)

	t.Run("deployment opens pods", func(t *testing.T) {
		w, _ := setupWorkloadView(t, dp)
		w.showRes(w.App(), w.GetTable().GetModel(), client.WkGVR, "apps/v1/deployments|default|web")

		top := w.App().Content.Top()
		require.NotNil(t, top)
		po, ok := top.(*Pod)
		require.True(t, ok, "expected Pod view, got %T", top)
		assert.Equal(t, "app=web", po.GetTable().GetModel().GetLabelSelector().String())
	})

	t.Run("statefulset opens pods", func(t *testing.T) {
		w, _ := setupWorkloadView(t, sts)
		w.showRes(w.App(), w.GetTable().GetModel(), client.WkGVR, "apps/v1/statefulsets|default|db")

		top := w.App().Content.Top()
		require.NotNil(t, top)
		po, ok := top.(*Pod)
		require.True(t, ok, "expected Pod view, got %T", top)
		assert.Equal(t, "app=db", po.GetTable().GetModel().GetLabelSelector().String())
	})

	t.Run("other kind keeps gotoResource behavior", func(t *testing.T) {
		w, _ := setupWorkloadView(t)
		w.showRes(w.App(), w.GetTable().GetModel(), client.WkGVR, "v1/services|default|svc1")

		top := w.App().Content.Top()
		require.NotNil(t, top)
		_, isPod := top.(*Pod)
		assert.False(t, isPod)
		_, isSvc := top.(*Service)
		assert.True(t, isSvc, "expected Service view, got %T", top)
	})
}

func TestWorkloadReplicaSetsCmd(t *testing.T) {
	replicas := int32(2)
	dp := makeTestDeployment("web", &replicas)
	w, _ := setupWorkloadView(t, dp)
	selectWorkloadRow(w, "apps/v1/deployments|default|web")

	evt := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)
	assert.Nil(t, w.replicaSetsCmd(evt))

	top := w.App().Content.Top()
	require.NotNil(t, top)
	rs, ok := top.(*ReplicaSet)
	require.True(t, ok, "expected ReplicaSet view, got %T", top)
	assert.Equal(t, "app=web", rs.GetTable().GetModel().GetLabelSelector().String())
}

func TestWorkloadLogOptionsAndCurrentReplicas(t *testing.T) {
	r3 := int32(3)
	r0 := int32(0)
	dp := makeTestDeployment("web", &r3)
	dpNil := makeTestDeployment("web-nil", nil)
	sts := makeTestStatefulSet("db", &r0)
	stsNil := makeTestStatefulSet("db-nil", nil)

	w, _ := setupWorkloadView(t, dp, dpNil, sts, stsNil)

	opts, err := w.logOptions(workloadTarget{gvr: client.DpGVR, fqn: "default/web"}, true)
	require.NoError(t, err)
	assert.Equal(t, "default/web", opts.Path)
	assert.True(t, opts.Previous)
	assert.Equal(t, "app", opts.Container)

	opts, err = w.logOptions(workloadTarget{gvr: client.StsGVR, fqn: "default/db"}, false)
	require.NoError(t, err)
	assert.Equal(t, "default/db", opts.Path)
	assert.False(t, opts.Previous)
	assert.Equal(t, "db", opts.Container)

	got, err := w.currentReplicas(workloadTarget{gvr: client.DpGVR, fqn: "default/web"})
	require.NoError(t, err)
	assert.Equal(t, "3", got)

	got, err = w.currentReplicas(workloadTarget{gvr: client.DpGVR, fqn: "default/web-nil"})
	require.NoError(t, err)
	assert.Equal(t, "1", got)

	got, err = w.currentReplicas(workloadTarget{gvr: client.StsGVR, fqn: "default/db"})
	require.NoError(t, err)
	assert.Equal(t, "0", got)

	got, err = w.currentReplicas(workloadTarget{gvr: client.StsGVR, fqn: "default/db-nil"})
	require.NoError(t, err)
	assert.Equal(t, "1", got)
}

func TestWorkloadScaleRestartAndSetImage(t *testing.T) {
	r2 := int32(2)
	dp := makeTestDeployment("web", &r2)
	sts := makeTestStatefulSet("db", &r2)
	w, kubeClient := setupWorkloadView(t, dp, sts)

	var scaledReplicas map[string]int32 = make(map[string]int32)
	kubeClient.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "scale" {
			return true, &autoscalingv1.Scale{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web"},
				Spec:       autoscalingv1.ScaleSpec{Replicas: 2},
			}, nil
		}
		return false, nil, nil
	})
	kubeClient.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "scale" {
			s := action.(k8stesting.UpdateAction).GetObject().(*autoscalingv1.Scale)
			scaledReplicas["deployments/web"] = s.Spec.Replicas
			return true, s, nil
		}
		return false, nil, nil
	})
	kubeClient.PrependReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "scale" {
			return true, &autoscalingv1.Scale{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "db"},
				Spec:       autoscalingv1.ScaleSpec{Replicas: 2},
			}, nil
		}
		return false, nil, nil
	})
	kubeClient.PrependReactor("update", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "scale" {
			s := action.(k8stesting.UpdateAction).GetObject().(*autoscalingv1.Scale)
			scaledReplicas["statefulsets/db"] = s.Spec.Replicas
			return true, s, nil
		}
		return false, nil, nil
	})

	ctx := context.Background()
	require.NoError(t, w.scaleTarget(ctx, workloadTarget{gvr: client.DpGVR, fqn: "default/web"}, 5))
	require.NoError(t, w.scaleTarget(ctx, workloadTarget{gvr: client.StsGVR, fqn: "default/db"}, 4))
	assert.Equal(t, int32(5), scaledReplicas["deployments/web"])
	assert.Equal(t, int32(4), scaledReplicas["statefulsets/db"])

	require.NoError(t, w.restartRollout(ctx, workloadTarget{gvr: client.DpGVR, fqn: "default/web"}, &metav1.PatchOptions{}))
	require.NoError(t, w.restartRollout(ctx, workloadTarget{gvr: client.StsGVR, fqn: "default/db"}, &metav1.PatchOptions{}))

	dpSpec, err := w.getPodSpec(workloadTarget{gvr: client.DpGVR, fqn: "default/web"})
	require.NoError(t, err)
	assert.Equal(t, "nginx:1.25", dpSpec.Containers[0].Image)

	stsSpec, err := w.getPodSpec(workloadTarget{gvr: client.StsGVR, fqn: "default/db"})
	require.NoError(t, err)
	assert.Equal(t, "postgres:16", stsSpec.Containers[0].Image)

	require.NoError(t, w.setImages(ctx, workloadTarget{gvr: client.DpGVR, fqn: "default/web"}, dao.ImageSpecs{{Name: "app", DockerImage: "nginx:1.27"}}))
	require.NoError(t, w.setImages(ctx, workloadTarget{gvr: client.StsGVR, fqn: "default/db"}, dao.ImageSpecs{{Name: "db", DockerImage: "postgres:17"}}))
}

func TestWorkloadPortForwardContext(t *testing.T) {
	w, _ := setupWorkloadView(t)
	w.App().BenchFile = "/tmp/bench.yaml"

	ctx := w.portForwardContext("default/web")(context.Background())
	assert.Equal(t, "default/web", ctx.Value(internal.KeyPath))
	assert.Equal(t, "/tmp/bench.yaml", ctx.Value(internal.KeyBenchCfg))
}

func TestWorkloadSelectedTargetsFiltersUnsupported(t *testing.T) {
	w, _ := setupWorkloadView(t)
	for _, id := range []string{
		"apps/v1/deployments|default|web",
		"v1/pods|default|p1",
		"apps/v1/statefulsets|default|db",
		"v1/services|default|svc1",
	} {
		selectWorkloadRow(w, id)
		w.GetTable().ToggleMark()
	}

	targets := w.selectedTargets()
	assert.ElementsMatch(t, []workloadTarget{
		{gvr: client.DpGVR, fqn: "default/web"},
		{gvr: client.StsGVR, fqn: "default/db"},
	}, targets)
}

func TestWorkloadLogsCmdOpensLogView(t *testing.T) {
	r1 := int32(1)
	dp := makeTestDeployment("web", &r1)
	sts := makeTestStatefulSet("db", &r1)

	for _, tc := range []struct {
		name  string
		rowID string
		prev  bool
	}{
		{name: "deployment-logs", rowID: "apps/v1/deployments|default|web", prev: false},
		{name: "deployment-prev-logs", rowID: "apps/v1/deployments|default|web", prev: true},
		{name: "statefulset-logs", rowID: "apps/v1/statefulsets|default|db", prev: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, _ := setupWorkloadView(t, dp, sts)
			selectWorkloadRow(w, tc.rowID)

			evt := tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone)
			assert.Nil(t, w.logsCmd(tc.prev)(evt))

			top := w.App().Content.Top()
			require.NotNil(t, top)
			_, ok := top.(*Log)
			assert.True(t, ok, "expected Log view, got %T", top)
		})
	}
}

func TestWorkloadFetchPodNameAndShowPFCmd(t *testing.T) {
	r1 := int32(1)
	dp := makeTestDeployment("web", &r1)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "web-abc-123",
			Labels:    map[string]string{"app": "web"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	w, _ := setupWorkloadView(t, dp, pod)
	selectWorkloadRow(w, "apps/v1/deployments|default|web")

	podName, err := w.fetchPodName(workloadTarget{gvr: client.DpGVR, fqn: "default/web"})
	require.NoError(t, err)
	assert.Equal(t, "default/web-abc-123", podName)

	// Without an active port-forward, showPFCmd should consume the key without pushing a PortForward view.
	evt := tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone)
	assert.Nil(t, w.showPFCmd(evt))
	assert.Nil(t, w.App().Content.Top())
}

func TestWorkloadScaleAndSetImageForms(t *testing.T) {
	r3 := int32(3)
	dp := makeTestDeployment("web", &r3)
	dp.Spec.Template.Spec.InitContainers = []corev1.Container{
		{Name: "init-db", Image: "busybox:1.36"},
	}
	sts := makeTestStatefulSet("db", &r3)
	w, _ := setupWorkloadView(t, dp, sts)

	singleForm, err := w.makeScaleForm([]workloadTarget{{gvr: client.DpGVR, fqn: "default/web"}})
	require.NoError(t, err)
	replicasField, ok := singleForm.GetFormItemByLabel("Replicas:").(*tview.InputField)
	require.True(t, ok)
	assert.Equal(t, "3", replicasField.GetText())

	multiForm, err := w.makeScaleForm([]workloadTarget{
		{gvr: client.DpGVR, fqn: "default/web"},
		{gvr: client.StsGVR, fqn: "default/db"},
	})
	require.NoError(t, err)
	multiField, ok := multiForm.GetFormItemByLabel("Replicas:").(*tview.InputField)
	require.True(t, ok)
	assert.Equal(t, "0", multiField.GetText())

	imgForm, err := w.makeSetImageForm(workloadTarget{gvr: client.DpGVR, fqn: "default/web"})
	require.NoError(t, err)
	initField, ok := imgForm.GetFormItemByLabel("init-db").(*tview.InputField)
	require.True(t, ok)
	assert.Equal(t, "busybox:1.36", initField.GetText())
	appField, ok := imgForm.GetFormItemByLabel("app").(*tview.InputField)
	require.True(t, ok)
	assert.Equal(t, "nginx:1.25", appField.GetText())
}
