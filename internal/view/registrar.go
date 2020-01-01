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
	vv[client.NewGVR("v1/namespaces")] = MetaViewer{
		viewerFn: NewNamespace,
	}
	vv[client.NewGVR("v1/events")] = MetaViewer{
		viewerFn: NewEvent,
	}
	vv[client.NewGVR("v1/pods")] = MetaViewer{
		viewerFn: NewPod,
	}
	vv[client.NewGVR("v1/services")] = MetaViewer{
		viewerFn: NewService,
	}
	vv[client.NewGVR("v1/nodes")] = MetaViewer{
		viewerFn: NewNode,
	}
	vv[client.NewGVR("v1/secrets")] = MetaViewer{
		viewerFn: NewSecret,
	}
}

func miscRes(vv MetaViewers) {
	vv[client.NewGVR("contexts")] = MetaViewer{
		viewerFn: NewContext,
	}
	vv[client.NewGVR("containers")] = MetaViewer{
		viewerFn: NewContainer,
	}
	vv[client.NewGVR("portforwards")] = MetaViewer{
		viewerFn: NewPortForward,
	}
	vv[client.NewGVR("screendumps")] = MetaViewer{
		viewerFn: NewScreenDump,
	}
	vv[client.NewGVR("benchmarks")] = MetaViewer{
		viewerFn: NewBenchmark,
	}
	vv[client.NewGVR("aliases")] = MetaViewer{
		viewerFn: NewAlias,
	}
}

func appsRes(vv MetaViewers) {
	vv[client.NewGVR("apps/v1/deployments")] = MetaViewer{
		viewerFn: NewDeploy,
	}
	vv[client.NewGVR("apps/v1/replicasets")] = MetaViewer{
		viewerFn: NewReplicaSet,
	}
	vv[client.NewGVR("apps/v1/statefulsets")] = MetaViewer{
		viewerFn: NewStatefulSet,
	}
	vv[client.NewGVR("apps/v1/daemonsets")] = MetaViewer{
		viewerFn: NewDaemonSet,
	}
	vv[client.NewGVR("extensions/v1beta1/daemonsets")] = MetaViewer{
		viewerFn: NewDaemonSet,
	}
}

func rbacRes(vv MetaViewers) {
	vv[client.NewGVR("rbac")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("users")] = MetaViewer{
		viewerFn: NewUser,
	}
	vv[client.NewGVR("groups")] = MetaViewer{
		viewerFn: NewGroup,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/clusterroles")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/roles")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/clusterrolebindings")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/rolebindings")] = MetaViewer{
		enterFn: showRules,
	}
}

func batchRes(vv MetaViewers) {
	vv[client.NewGVR("batch/v1beta1/cronjobs")] = MetaViewer{
		viewerFn: NewCronJob,
	}
	vv[client.NewGVR("batch/v1/jobs")] = MetaViewer{
		viewerFn: NewJob,
	}
}

func extRes(vv MetaViewers) {
	vv[client.NewGVR("apiextensions.k8s.io/v1/customresourcedefinitions")] = MetaViewer{
		enterFn: showCRD,
	}
	vv[client.NewGVR("apiextensions.k8s.io/v1beta1/customresourcedefinitions")] = MetaViewer{
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
