package view

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var aliases = config.NewAliases()

func resourceFn(l resource.List) ViewFunc {
	return func(title, gvr string, list resource.List) ResourceViewer {
		return NewResource(title, gvr, l)
	}
}

func allCRDs(c k8s.Connection, vv MetaViewers) {
	crds, err := resource.NewCustomResourceDefinitionList(c, resource.AllNamespaces).
		Resource().
		List(resource.AllNamespaces, metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("CRDs load fail")
		return
	}

	for _, crd := range crds {
		meta, err := crd.ExtFields()
		if err != nil {
			log.Error().Err(err).Msgf("Error getting extended fields from %s", crd.Name())
			continue
		}

		gvr := k8s.NewGVR(meta.Group, meta.Version, meta.Plural)
		gvrs := gvr.String()
		if meta.Plural != "" {
			aliases.Define(gvrs, meta.Plural)
		}
		if meta.Singular != "" {
			aliases.Define(gvrs, meta.Singular)
		}
		for _, a := range meta.ShortNames {
			aliases.Define(gvrs, a)
		}

		vv[gvrs] = MetaViewer{
			gvr:       gvrs,
			kind:      meta.Kind,
			viewFn:    resourceFn(resource.NewCustomList(c, meta.Namespaced, "", gvrs)),
			colorerFn: ui.DefaultColorer,
		}
	}
}

func showRBAC(app *App, ns, resource, selection string) {
	kind := ClusterRole
	if resource == "role" {
		kind = Role
	}
	app.inject(NewRbac(selection, kind))
}

func showCRD(app *App, ns, resource, selection string) {
	log.Debug().Msgf("Launching CRD %q -- %q -- %q", ns, resource, selection)
	tokens := strings.Split(selection, ".")
	if !app.gotoResource(tokens[0]) {
		app.Flash().Errf("Goto %s failed", tokens[0])
	}
}

func showClusterRole(app *App, ns, resource, selection string) {
	crb, err := app.Conn().DialOrDie().RbacV1().ClusterRoleBindings().Get(selection, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve clusterrolebindings for %s", selection)
		return
	}
	app.inject(NewRbac(crb.RoleRef.Name, ClusterRole))
}

func showRole(app *App, _, resource, selection string) {
	ns, n := namespaced(selection)
	rb, err := app.Conn().DialOrDie().RbacV1().RoleBindings(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve rolebindings for %s", selection)
		return
	}
	app.inject(NewRbac(fqn(ns, rb.RoleRef.Name), Role))
}

func showSAPolicy(app *App, _, _, selection string) {
	_, n := namespaced(selection)
	subject, err := mapFuSubject("ServiceAccount")
	if err != nil {
		app.Flash().Err(err)
		return
	}
	app.inject(NewPolicy(app, subject, n))
}

func load(c k8s.Connection, vv MetaViewers) {
	if err := aliases.Load(); err != nil {
		log.Error().Err(err).Msg("No custom aliases defined in config")
	}
	discovery, err := c.CachedDiscovery()
	if err != nil {
		log.Error().Err(err).Msgf("Error to get discovery client")
		return
	}

	rr, _ := discovery.ServerPreferredResources()
	for _, r := range rr {
		for _, res := range r.APIResources {
			gvr := k8s.ToGVR(r.GroupVersion, res.Name)
			cmd, ok := vv[gvr.String()]
			if !ok {
				// log.Debug().Msgf(fmt.Sprintf(">> No viewer defined for `%s`", gvr))
				continue
			}
			cmd.namespaced = res.Namespaced
			cmd.kind = res.Kind
			cmd.verbs = res.Verbs
			cmd.gvr = gvr.String()
			vv[gvr.String()] = cmd
			gvrStr := gvr.String()
			aliases.Define(gvrStr, strings.ToLower(res.Kind))
			aliases.Define(gvrStr, res.Name)
			if len(res.SingularName) > 0 {
				aliases.Define(gvrStr, res.SingularName)
			}
			for _, s := range res.ShortNames {
				aliases.Define(gvrStr, s)
			}
		}
	}
}

func resourceViews(c k8s.Connection, m MetaViewers) {
	coreRes(m)
	miscRes(m)
	appsRes(m)
	authRes(m)
	extRes(m)
	netRes(m)
	batchRes(m)
	policyRes(m)
	hpaRes(m)

	load(c, m)
}

func coreRes(vv MetaViewers) {
	vv["v1/nodes"] = MetaViewer{
		viewFn:    NewNode,
		listFn:    resource.NewNodeList,
		colorerFn: nsColorer,
	}
	vv["v1/namespaces"] = MetaViewer{
		viewFn:    NewNamespace,
		listFn:    resource.NewNamespaceList,
		colorerFn: nsColorer,
	}
	vv["v1/pods"] = MetaViewer{
		viewFn:    NewPod,
		listFn:    resource.NewPodList,
		colorerFn: podColorer,
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
		listFn:    resource.NewPersistentVolumeList,
		colorerFn: pvColorer,
	}
	vv["v1/persistentvolumeclaims"] = MetaViewer{
		listFn:    resource.NewPersistentVolumeClaimList,
		colorerFn: pvcColorer,
	}
	vv["v1/secrets"] = MetaViewer{
		viewFn: NewSecret,
		listFn: resource.NewSecretList,
	}
	vv["v1/endpoints"] = MetaViewer{
		listFn: resource.NewEndpointsList,
	}
	vv["v1/events"] = MetaViewer{
		listFn:    resource.NewEventList,
		colorerFn: evColorer,
	}
	vv["v1/replicationcontrollers"] = MetaViewer{
		viewFn:    NewReplicationController,
		listFn:    resource.NewReplicationControllerList,
		colorerFn: rsColorer,
	}
}

func miscRes(vv MetaViewers) {
	vv["storage.k8s.io/v1/storageclasses"] = MetaViewer{
		listFn: resource.NewStorageClassList,
	}
	vv["contexts"] = MetaViewer{
		gvr:       "contexts",
		kind:      "Contexts",
		viewFn:    NewContext,
		listFn:    resource.NewContextList,
		colorerFn: ctxColorer,
	}
	vv["users"] = MetaViewer{
		gvr:    "users",
		viewFn: NewSubject,
	}
	vv["groups"] = MetaViewer{
		gvr:    "groups",
		viewFn: NewSubject,
	}
	vv["portforwards"] = MetaViewer{
		gvr:    "portforwards",
		viewFn: NewPortForward,
	}
	vv["benchmarks"] = MetaViewer{
		gvr:    "benchmarks",
		viewFn: NewBench,
	}
	vv["screendumps"] = MetaViewer{
		gvr:    "screendumps",
		viewFn: NewScreenDump,
	}
}

func appsRes(vv MetaViewers) {
	vv["apps/v1/deployments"] = MetaViewer{
		viewFn:    NewDeploy,
		listFn:    resource.NewDeploymentList,
		colorerFn: dpColorer,
	}
	vv["apps/v1/replicasets"] = MetaViewer{
		viewFn:    NewReplicaSet,
		listFn:    resource.NewReplicaSetList,
		colorerFn: rsColorer,
	}
	vv["apps/v1/statefulsets"] = MetaViewer{
		viewFn:    NewStatefulSet,
		listFn:    resource.NewStatefulSetList,
		colorerFn: stsColorer,
	}
	vv["apps/v1/daemonsets"] = MetaViewer{
		viewFn:    NewDaemonSet,
		listFn:    resource.NewDaemonSetList,
		colorerFn: dpColorer,
	}
}

func authRes(vv MetaViewers) {
	vv["rbac.authorization.k8s.io/v1/clusterroles"] = MetaViewer{
		listFn:  resource.NewClusterRoleList,
		enterFn: showRBAC,
	}
	vv["rbac.authorization.k8s.io/v1/clusterrolebindings"] = MetaViewer{
		listFn:  resource.NewClusterRoleBindingList,
		enterFn: showClusterRole,
	}
	vv["rbac.authorization.k8s.io/v1/rolebindings"] = MetaViewer{
		listFn:  resource.NewRoleBindingList,
		enterFn: showRole,
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
		listFn:    resource.NewPDBList,
		colorerFn: pdbColorer,
	}
}

func hpaRes(vv MetaViewers) {
	vv["autoscaling/v1/horizontalpodautoscalers"] = MetaViewer{
		listFn: resource.NewHorizontalPodAutoscalerV1List,
	}
}
