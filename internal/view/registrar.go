package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
)

func loadCustomViewers() MetaViewers {
	m := make(MetaViewers, 30)
	coreRes(m)
	miscRes(m)
	appsRes(m)
	rbacRes(m)
	batchRes(m)
	extRes(m)

	return m
}

func coreRes(vv MetaViewers) {
	vv["v1/namespaces"] = MetaViewer{
		viewerFn: NewNamespace,
	}
	vv["v1/events"] = MetaViewer{
		viewerFn: NewEvent,
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
		enterFn: showRules,
	}
	vv["users"] = MetaViewer{
		viewerFn: NewUser,
	}
	vv["groups"] = MetaViewer{
		viewerFn: NewGroup,
	}
	vv["rbac.authorization.k8s.io/v1/clusterroles"] = MetaViewer{
		enterFn: showRules,
	}
	vv["rbac.authorization.k8s.io/v1/roles"] = MetaViewer{
		enterFn: showRules,
	}
	vv["rbac.authorization.k8s.io/v1/clusterrolebindings"] = MetaViewer{
		enterFn: showRules,
	}
	vv["rbac.authorization.k8s.io/v1/rolebindings"] = MetaViewer{
		enterFn: showRules,
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

func extRes(vv MetaViewers) {
	vv["apiextensions.k8s.io/v1/customresourcedefinitions"] = MetaViewer{
		enterFn: showCRD,
	}
	vv["apiextensions.k8s.io/v1beta1/customresourcedefinitions"] = MetaViewer{
		enterFn: showCRD,
	}
}

func showCRD(app *App, ns, gvr, path string) {
	_, crdGVR := client.Namespaced(path)
	tokens := strings.Split(crdGVR, ".")
	if err := app.gotoResource(tokens[0], false); err != nil {
		app.Flash().Err(err)
	}
}
