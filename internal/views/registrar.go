package views

import (
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	viewFn     func(title, gvr string, app *appView, list resource.List) resourceViewer
	listFn     func(c resource.Connection, ns string) resource.List
	enterFn    func(app *appView, ns, resource, selection string)
	decorateFn func(resource.TableData) resource.TableData

	viewer struct {
		gvr        string
		kind       string
		namespaced bool
		verbs      metav1.Verbs
		viewFn     viewFn
		listFn     listFn
		enterFn    enterFn
		colorerFn  ui.ColorerFunc
		decorateFn decorateFn
	}

	viewers map[string]viewer
)

func listFunc(l resource.List) viewFn {
	return func(title, gvr string, app *appView, list resource.List) resourceViewer {
		return newResourceView(title, gvr, app, l)
	}
}

var aliases = config.NewAliases()

func allCRDs(c k8s.Connection, vv viewers) {
	crds, err := resource.NewCustomResourceDefinitionList(c, resource.AllNamespaces).
		Resource().
		List(resource.AllNamespaces)
	if err != nil {
		log.Error().Err(err).Msg("CRDs load fail")
		return
	}

	t := time.Now()
	for _, crd := range crds {
		meta, err := crd.ExtFields()
		if err != nil {
			log.Error().Err(err).Msgf("Error getting extended fields from %s", crd.Name())
			continue
		}

		gvr := k8s.NewGVR(meta.Group, meta.Version, meta.Plural)
		gvrs := gvr.String()
		if meta.Plural != "" {
			aliases.Define(meta.Plural, gvrs)
		}
		if meta.Singular != "" {
			aliases.Define(meta.Singular, gvrs)
		}
		for _, a := range meta.ShortNames {
			aliases.Define(a, gvrs)
		}

		vv[gvrs] = viewer{
			gvr:       gvrs,
			kind:      meta.Kind,
			viewFn:    listFunc(resource.NewCustomList(c, meta.Namespaced, "", gvrs)),
			colorerFn: ui.DefaultColorer,
		}
	}
	log.Debug().Msgf("Loading CRDS %v", time.Since(t))
}

func showRBAC(app *appView, ns, resource, selection string) {
	kind := clusterRole
	if resource == "role" {
		kind = role
	}
	app.inject(newRBACView(app, ns, selection, kind))
}

func showCRD(app *appView, ns, resource, selection string) {
	log.Debug().Msgf("Launching CRD %q -- %q -- %q", ns, resource, selection)
	tokens := strings.Split(selection, ".")
	app.gotoResource(tokens[0], true)

}

func showClusterRole(app *appView, ns, resource, selection string) {
	crb, err := app.Conn().DialOrDie().RbacV1().ClusterRoleBindings().Get(selection, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve clusterrolebindings for %s", selection)
		return
	}
	app.inject(newRBACView(app, ns, crb.RoleRef.Name, clusterRole))
}

func showRole(app *appView, _, resource, selection string) {
	ns, n := namespaced(selection)
	rb, err := app.Conn().DialOrDie().RbacV1().RoleBindings(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve rolebindings for %s", selection)
		return
	}
	app.inject(newRBACView(app, ns, fqn(ns, rb.RoleRef.Name), role))
}

func showSAPolicy(app *appView, _, _, selection string) {
	_, n := namespaced(selection)
	app.inject(newPolicyView(app, mapFuSubject("ServiceAccount"), n))
}

func load(c k8s.Connection, vv viewers) {
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
			aliases.Define(strings.ToLower(res.Kind), gvrStr)
			aliases.Define(res.Name, gvrStr)
			if len(res.SingularName) > 0 {
				aliases.Define(res.SingularName, gvrStr)
			}
			for _, s := range res.ShortNames {
				aliases.Define(s, gvrStr)
			}
		}
	}
}

func resourceViews(c k8s.Connection, m viewers) {
	defer func(t time.Time) {
		log.Debug().Msgf("Loading Views Elapsed %v", time.Since(t))
	}(time.Now())

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

func coreRes(vv viewers) {
	vv["v1/nodes"] = viewer{
		viewFn:    newNodeView,
		listFn:    resource.NewNodeList,
		colorerFn: nsColorer,
	}
	vv["v1/namespaces"] = viewer{
		viewFn:    newNamespaceView,
		listFn:    resource.NewNamespaceList,
		colorerFn: nsColorer,
	}
	vv["v1/pods"] = viewer{
		viewFn:    newPodView,
		listFn:    resource.NewPodList,
		colorerFn: podColorer,
	}
	vv["v1/serviceaccounts"] = viewer{
		listFn:  resource.NewServiceAccountList,
		enterFn: showSAPolicy,
	}
	vv["v1/services"] = viewer{
		viewFn: newSvcView,
		listFn: resource.NewServiceList,
	}
	vv["v1/configmaps"] = viewer{
		listFn: resource.NewConfigMapList,
	}
	vv["v1/persistentvolumes"] = viewer{
		listFn:    resource.NewPersistentVolumeList,
		colorerFn: pvColorer,
	}
	vv["v1/persistentvolumeclaims"] = viewer{
		listFn:    resource.NewPersistentVolumeClaimList,
		colorerFn: pvcColorer,
	}
	vv["v1/secrets"] = viewer{
		viewFn: newSecretView,
		listFn: resource.NewSecretList,
	}
	vv["v1/endpoints"] = viewer{
		listFn: resource.NewEndpointsList,
	}
	vv["v1/events"] = viewer{
		listFn:    resource.NewEventList,
		colorerFn: evColorer,
	}
	vv["v1/replicationcontrollers"] = viewer{
		viewFn:    newScalableResourceView,
		listFn:    resource.NewReplicationControllerList,
		colorerFn: rsColorer,
	}
}

func miscRes(vv viewers) {
	vv["storage.k8s.io/v1/storageclasses"] = viewer{
		listFn: resource.NewStorageClassList,
	}
	vv["contexts"] = viewer{
		gvr:       "contexts",
		kind:      "Contexts",
		viewFn:    newContextView,
		listFn:    resource.NewContextList,
		colorerFn: ctxColorer,
	}
	vv["users"] = viewer{
		gvr:    "users",
		viewFn: newSubjectView,
	}
	vv["groups"] = viewer{
		gvr:    "groups",
		viewFn: newSubjectView,
	}
	vv["portforwards"] = viewer{
		gvr:    "portforwards",
		viewFn: newForwardView,
	}
	vv["benchmarks"] = viewer{
		gvr:    "benchmarks",
		viewFn: newBenchView,
	}
	vv["screendumps"] = viewer{
		gvr:    "screendumps",
		viewFn: newDumpView,
	}
}

func appsRes(vv viewers) {
	vv["apps/v1/deployments"] = viewer{
		viewFn:    newDeployView,
		listFn:    resource.NewDeploymentList,
		colorerFn: dpColorer,
	}
	vv["apps/v1/replicasets"] = viewer{
		viewFn:    newReplicaSetView,
		listFn:    resource.NewReplicaSetList,
		colorerFn: rsColorer,
	}
	vv["apps/v1/statefulsets"] = viewer{
		viewFn:    newStatefulSetView,
		listFn:    resource.NewStatefulSetList,
		colorerFn: stsColorer,
	}
	vv["apps/v1/daemonsets"] = viewer{
		viewFn:    newDaemonSetView,
		listFn:    resource.NewDaemonSetList,
		colorerFn: dpColorer,
	}
}

func authRes(vv viewers) {
	vv["rbac.authorization.k8s.io/v1/clusterroles"] = viewer{
		listFn:  resource.NewClusterRoleList,
		enterFn: showRBAC,
	}
	vv["rbac.authorization.k8s.io/v1/clusterrolebindings"] = viewer{
		listFn:  resource.NewClusterRoleBindingList,
		enterFn: showClusterRole,
	}
	vv["rbac.authorization.k8s.io/v1/rolebindings"] = viewer{
		listFn:  resource.NewRoleBindingList,
		enterFn: showRole,
	}
	vv["rbac.authorization.k8s.io/v1/roles"] = viewer{
		listFn:  resource.NewRoleList,
		enterFn: showRBAC,
	}
}

func extRes(vv viewers) {
	vv["apiextensions.k8s.io/v1/customresourcedefinitions"] = viewer{
		listFn:  resource.NewCustomResourceDefinitionList,
		enterFn: showCRD,
	}
}

func netRes(vv viewers) {
	vv["networking.k8s.io/v1/networkpolicies"] = viewer{
		listFn: resource.NewNetworkPolicyList,
	}
	vv["extensions/v1beta1/ingresses"] = viewer{
		listFn: resource.NewIngressList,
	}
}

func batchRes(vv viewers) {
	vv["batch/v1beta1/cronjobs"] = viewer{
		viewFn: newCronJobView,
		listFn: resource.NewCronJobList,
	}
	vv["batch/v1/jobs"] = viewer{
		viewFn: newJobView,
		listFn: resource.NewJobList,
	}
}

func policyRes(vv viewers) {
	vv["policy/v1beta1/poddisruptionbudgets"] = viewer{
		listFn:    resource.NewPDBList,
		colorerFn: pdbColorer,
	}
}

func hpaRes(vv viewers) {
	vv["autoscaling/v1/horizontalpodautoscalers"] = viewer{
		listFn: resource.NewHorizontalPodAutoscalerV1List,
	}
}
