package view

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var aliases = config.NewAliases()

func ToResource(o *unstructured.Unstructured, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, &obj)
}

func loadAliases() error {
	if err := aliases.Load(); err != nil {
		return err
	}
	for _, gvr := range dao.AllGVRs() {
		meta, err := dao.MetaFor(gvr)
		if err != nil {
			return err
		}
		if _, ok := aliases.Alias[meta.Kind]; ok {
			continue
		}
		aliases.Define(string(gvr), strings.ToLower(meta.Kind), meta.Name)
		if meta.SingularName != "" {
			aliases.Define(string(gvr), meta.SingularName)
		}
		if meta.ShortNames != nil {
			aliases.Define(string(gvr), meta.ShortNames...)
		}
	}

	return nil
}

func loadCustomViewers() MetaViewers {
	m := make(MetaViewers, 30)

	coreRes(m)
	miscRes(m)
	appsRes(m)
	rbacRes(m)
	batchRes(m)

	return m
}

func coreRes(vv MetaViewers) {
	vv["v1/namespaces"] = MetaViewer{
		viewerFn: NewNamespace,
	}
	vv["v1/pods"] = MetaViewer{
		viewerFn: NewPod,
	}
	vv["v1/services"] = MetaViewer{
		viewerFn: NewService,
	}
	vv["v1/nodes"] = MetaViewer{
		viewerFn: NewNode,
	}
	vv["v1/secrets"] = MetaViewer{
		viewerFn: NewSecret,
	}
}

func miscRes(vv MetaViewers) {
	vv["contexts"] = MetaViewer{
		viewerFn: NewContext,
	}
	vv["containers"] = MetaViewer{
		viewerFn: NewContainer,
	}
	vv["portforwards"] = MetaViewer{
		viewerFn: NewPortForward,
	}
	vv["screendumps"] = MetaViewer{
		viewerFn: NewScreenDump,
	}
	vv["benchmarks"] = MetaViewer{
		viewerFn: NewBenchmark,
	}
	vv["aliases"] = MetaViewer{
		viewerFn: NewAlias,
	}
}

func appsRes(vv MetaViewers) {
	vv["apps/v1/deployments"] = MetaViewer{
		viewerFn: NewDeploy,
	}
	vv["apps/v1/replicasets"] = MetaViewer{
		viewerFn: NewReplicaSet,
	}
	vv["apps/v1/statefulsets"] = MetaViewer{
		viewerFn: NewStatefulSet,
	}
	vv["apps/v1/daemonsets"] = MetaViewer{
		viewerFn: NewDaemonSet,
	}
	vv["extensions/v1beta1/daemonsets"] = MetaViewer{
		viewerFn: NewDaemonSet,
	}
}

func rbacRes(vv MetaViewers) {
	vv["rbac"] = MetaViewer{
		enterFn: showRBAC,
	}
	vv["rbac.authorization.k8s.io/v1/clusterroles"] = MetaViewer{
		enterFn: showRBAC,
	}
	vv["rbac.authorization.k8s.io/v1/roles"] = MetaViewer{
		enterFn: showRBAC,
	}
	vv["rbac.authorization.k8s.io/v1/clusterrolebindings"] = MetaViewer{
		enterFn: showRBAC,
	}
	vv["rbac.authorization.k8s.io/v1/rolebindings"] = MetaViewer{
		enterFn: showRBAC,
	}
}

func batchRes(vv MetaViewers) {
	vv["batch/v1beta1/cronjobs"] = MetaViewer{
		viewerFn: NewCronJob,
	}
	vv["batch/v1/jobs"] = MetaViewer{
		viewerFn: NewJob,
	}
}
