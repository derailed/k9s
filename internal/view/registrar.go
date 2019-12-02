package view

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var aliases = config.NewAliases()

func resourceFn(l resource.List) ViewFunc {
	return func(title, gvr string, list resource.List) ResourceViewer {
		return NewResource(title, gvr, l)
	}
}

func ToResource(o *unstructured.Unstructured, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, &obj)
}

func showCRD(app *App, ns, resource, selection string) {
	log.Debug().Msgf("Launching CRD %q -- %q -- %q", ns, resource, selection)
	tokens := strings.Split(selection, ".")
	_ = tokens
	panic("NYI")
	// if !app.gotoResource(tokens[0]) {
	// 	app.Flash().Errf("Goto %s failed", tokens[0])
	// }
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
	authRes(m)
	extRes(m)
	netRes(m)
	batchRes(m)
	policyRes(m)
	hpaRes(m)

	return m
}

func coreRes(vv MetaViewers) {
	vv["v1/nodes"] = MetaViewer{
		viewFn: NewNode,
		listFn: resource.NewNodeList,
	}
	vv["v1/namespaces"] = MetaViewer{
		viewerFn: NewNamespace,
	}
	vv["v1/pods"] = MetaViewer{
		viewFn: NewPod,
		listFn: resource.NewPodList,
	}
	vv["v1/serviceaccounts"] = MetaViewer{
		listFn:  resource.NewServiceAccountList,
		enterFn: showSAPolicy,
	}
	vv["v1/services"] = MetaViewer{
		viewFn: NewService,
		listFn: resource.NewServiceList,
	}
	vv["v1/configmaps"] = MetaViewer{
		listFn: resource.NewConfigMapList,
	}
	vv["v1/persistentvolumes"] = MetaViewer{
		listFn: resource.NewPersistentVolumeList,
	}
	vv["v1/persistentvolumeclaims"] = MetaViewer{
		listFn: resource.NewPersistentVolumeClaimList,
	}
	vv["v1/secrets"] = MetaViewer{
		viewFn: NewSecret,
		listFn: resource.NewSecretList,
	}
	vv["v1/endpoints"] = MetaViewer{
		listFn: resource.NewEndpointsList,
	}
	vv["v1/events"] = MetaViewer{
		listFn: resource.NewEventList,
	}
	vv["v1/replicationcontrollers"] = MetaViewer{
		viewFn: NewReplicationController,
		listFn: resource.NewReplicationControllerList,
	}
}

func miscRes(vv MetaViewers) {
	vv["storage.k8s.io/v1/storageclasses"] = MetaViewer{
		listFn: resource.NewStorageClassList,
	}
	vv["contexts"] = MetaViewer{
		viewerFn: NewContext,
	}
	vv["users"] = MetaViewer{
		viewFn: NewSubject,
	}
	vv["groups"] = MetaViewer{
		viewFn: NewSubject,
	}
	vv["portforwards"] = MetaViewer{
		viewFn: NewPortForward,
	}
	vv["benchmarks"] = MetaViewer{
		viewFn: NewBench,
	}
	vv["screendumps"] = MetaViewer{
		viewerFn: NewScreenDump,
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

func authRes(vv MetaViewers) {
	vv["rbac.authorization.k8s.io/v1/clusterroles"] = MetaViewer{
		listFn:  resource.NewClusterRoleList,
		enterFn: showRBAC,
	}
	vv["rbac.authorization.k8s.io/v1/clusterrolebindings"] = MetaViewer{
		listFn:  resource.NewClusterRoleBindingList,
		enterFn: showClusterRoleBinding,
	}
	vv["rbac.authorization.k8s.io/v1/rolebindings"] = MetaViewer{
		listFn:  resource.NewRoleBindingList,
		enterFn: showRoleBinding,
	}
	vv["rbac.authorization.k8s.io/v1/roles"] = MetaViewer{
		listFn:  resource.NewRoleList,
		enterFn: showRBAC,
	}
}

func extRes(vv MetaViewers) {
	vv["apiextensions.k8s.io/v1/customresourcedefinitions"] = MetaViewer{
		listFn:  resource.NewCustomResourceDefinitionList,
		enterFn: showCRD,
	}
	vv["apiextensions.k8s.io/v1beta1/customresourcedefinitions"] = MetaViewer{
		listFn:  resource.NewCustomResourceDefinitionList,
		enterFn: showCRD,
	}
}

func netRes(vv MetaViewers) {
	vv["networking.k8s.io/v1/networkpolicies"] = MetaViewer{
		listFn: resource.NewNetworkPolicyList,
	}
	vv["extensions/v1beta1/ingresses"] = MetaViewer{
		listFn: resource.NewIngressList,
	}
}

func batchRes(vv MetaViewers) {
	vv["batch/v1beta1/cronjobs"] = MetaViewer{
		viewFn: NewCronJob,
		listFn: resource.NewCronJobList,
	}
	vv["batch/v1/jobs"] = MetaViewer{
		viewFn: NewJob,
		listFn: resource.NewJobList,
	}
}

func policyRes(vv MetaViewers) {
	vv["policy/v1beta1/poddisruptionbudgets"] = MetaViewer{
		listFn: resource.NewPDBList,
	}
}

func hpaRes(vv MetaViewers) {
	vv["autoscaling/v1/horizontalpodautoscalers"] = MetaViewer{
		listFn: resource.NewHorizontalPodAutoscalerV1List,
	}
}
