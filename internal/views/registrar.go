package views

import (
	"fmt"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	viewFn     func(ns string, app *appView, list resource.List) resourceViewer
	listFn     func(c resource.Connection, ns string) resource.List
	enterFn    func(app *appView, ns, resource, selection string)
	decorateFn func(resource.TableData) resource.TableData

	crdCmd struct {
		api      string
		version  string
		plural   string
		singular string
	}

	resCmd struct {
		crdCmd

		gvr        string
		namespaced bool
		verbs      metav1.Verbs
		viewFn     viewFn
		listFn     listFn
		enterFn    enterFn
		colorerFn  ui.ColorerFunc
		decorateFn decorateFn
	}

	AliasConfig struct {
		Aliases map[string]string `yaml:"aliases"`
	}
)

var DefaultAliasConfig = AliasConfig{
	Aliases: map[string]string{
		"Deployment": "dp",
		"Secret":     "sec",
		"Jobs":       "jo",

		"ClusterRoles":        "cr",
		"ClusterRoleBindings": "crb",
		"RoleBindings":        "rb",
		"Roles":               "ro",

		"NetworkPolicies": "np",

		"Contexts":    "ctx",
		"Users":       "usr",
		"Groups":      "grp",
		"PortForward": "pf",
		"Benchmark":   "be",
		"ScreenDumps": "sd",
	},
}

func aliasCmds(c k8s.Connection, m map[string]resCmd) {
	resourceViews(c, m)
	if c != nil {
		allCRDs(c, m)
	}
}

func listFunc(l resource.List) viewFn {
	return func(ns string, app *appView, list resource.List) resourceViewer {
		return newResourceView(
			ns,
			app,
			l,
		)
	}
}

func allCRDs(c k8s.Connection, m map[string]resCmd) {
	crds, err := resource.NewCustomResourceDefinitionList(c, resource.AllNamespaces).
		Resource().
		List(resource.AllNamespaces)
	if err != nil {
		log.Error().Err(err).Msg("CRDs load fail")
		return
	}

	for _, crd := range crds {
		ff := crd.ExtFields()

		gvr := path.Join(ff["group"].(string), ff["version"].(string), ff["kind"].(string))
		var name string
		if p, ok := ff["plural"].(string); ok {
			aliases[p] = gvr
			name = p
		}
		if s, ok := ff["singular"].(string); ok {
			aliases[s] = gvr
			name = s
		}
		if aa, ok := ff["aliases"].([]interface{}); ok {
			for _, a := range aa {
				aliases[a.(string)] = gvr
			}
		}
		m[gvr] = resCmd{
			gvr:       gvr,
			viewFn:    listFunc(resource.NewCustomList(c, "", ff["group"].(string), ff["version"].(string), name)),
			colorerFn: ui.DefaultColorer,
		}
	}
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

type Aliases map[string]string

var aliases Aliases

func load(c k8s.Connection, viewRes map[string]resCmd) {
	// cc := map[string]resCmd{}
	aliases = make(Aliases, len(viewRes))
	rr, _ := c.DialOrDie().Discovery().ServerPreferredResources()
	for _, r := range rr {
		log.Debug().Msgf("Group %#v", r.GroupVersion)
		for _, res := range r.APIResources {
			log.Debug().Msgf("\tRes %s -- %q:%q -- %+v", res.Name, res.Group, res.Version, res.ShortNames)
			gvr := path.Join(r.GroupVersion, res.Name)
			// Get singular, plural, shortname and to alias under gvr name
			if cmd, ok := viewRes[gvr]; !ok {
				log.Error().Msgf(fmt.Sprintf(">> Missed %s", gvr))
			} else {
				log.Debug().Msgf("Res %#v", res)
				cmd.namespaced = res.Namespaced
				cmd.verbs = res.Verbs
				cmd.gvr = gvr
				viewRes[gvr] = cmd
				aliases[strings.ToLower(res.Kind)] = gvr
				aliases[res.Name] = gvr
				aliases[res.SingularName] = gvr
				for _, s := range res.ShortNames {
					aliases[s] = gvr
				}
			}
		}
	}
}

func resourceViews(c k8s.Connection, m map[string]resCmd) {
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

func coreRes(m map[string]resCmd) {
	m["v1/nodes"] = resCmd{
		viewFn:    newNodeView,
		listFn:    resource.NewNodeList,
		colorerFn: nsColorer,
	}
	m["v1/namespaces"] = resCmd{
		viewFn:    newNamespaceView,
		listFn:    resource.NewNamespaceList,
		colorerFn: nsColorer,
	}
	m["v1/pods"] = resCmd{
		viewFn:    newPodView,
		listFn:    resource.NewPodList,
		colorerFn: podColorer,
	}
	m["v1/serviceaccounts"] = resCmd{
		viewFn:  newResourceView,
		listFn:  resource.NewServiceAccountList,
		enterFn: showSAPolicy,
	}
	m["v1/services"] = resCmd{
		viewFn: newSvcView,
		listFn: resource.NewServiceList,
	}
	m["v1/configmaps"] = resCmd{
		listFn: resource.NewConfigMapList,
	}
	m["v1/persistentvolumes"] = resCmd{
		listFn:    resource.NewPersistentVolumeList,
		colorerFn: pvColorer,
	}
	m["v1/persistentvolumeclaims"] = resCmd{
		listFn:    resource.NewPersistentVolumeClaimList,
		colorerFn: pvcColorer,
	}
	m["v1/secrets"] = resCmd{
		viewFn: newSecretView,
		listFn: resource.NewSecretList,
	}
	m["v1/endpoints"] = resCmd{
		listFn: resource.NewEndpointsList,
	}
	m["v1/events"] = resCmd{
		listFn:    resource.NewEventList,
		colorerFn: evColorer,
	}
	m["v1/replicationcontrollers"] = resCmd{
		viewFn:    newScalableResourceView,
		listFn:    resource.NewReplicationControllerList,
		colorerFn: rsColorer,
	}
}

func miscRes(m map[string]resCmd) {
	m["storage.k8s.io/storageclasses"] = resCmd{
		listFn: resource.NewStorageClassList,
	}
	m["ctx"] = resCmd{
		gvr:       "Contexts",
		viewFn:    newContextView,
		listFn:    resource.NewContextList,
		colorerFn: ctxColorer,
	}
	m["usr"] = resCmd{
		viewFn: newSubjectView,
	}
	m["grp"] = resCmd{
		viewFn: newSubjectView,
	}
	m["pf"] = resCmd{
		gvr:    "PortForward",
		viewFn: newForwardView,
	}
	m["be"] = resCmd{
		gvr:    "Benchmark",
		viewFn: newBenchView,
	}
	m["sd"] = resCmd{
		gvr:    "ScreenDumps",
		viewFn: newDumpView,
	}

}

func appsRes(m map[string]resCmd) {
	m["apps/v1/deployments"] = resCmd{
		viewFn:    newDeployView,
		listFn:    resource.NewDeploymentList,
		colorerFn: dpColorer,
	}
	m["apps/v1/replicasets"] = resCmd{
		viewFn:    newReplicaSetView,
		listFn:    resource.NewReplicaSetList,
		colorerFn: rsColorer,
	}
	m["apps/v1/statefulsets"] = resCmd{
		viewFn:    newStatefulSetView,
		listFn:    resource.NewStatefulSetList,
		colorerFn: stsColorer,
	}
	m["apps/v1/daemonsets"] = resCmd{
		viewFn:    newDaemonSetView,
		listFn:    resource.NewDaemonSetList,
		colorerFn: dpColorer,
	}
}

func authRes(m map[string]resCmd) {
	m["rbac.authorization.k8s.io/v1/clusterroles"] = resCmd{
		listFn:  resource.NewClusterRoleList,
		enterFn: showRBAC,
	}
	m["rbac.authorization.k8s.io/v1/clusterrolebindings"] = resCmd{
		listFn:  resource.NewClusterRoleBindingList,
		enterFn: showClusterRole,
	}
	m["rbac.authorization.k8s.io/v1/rolebindings"] = resCmd{
		listFn:  resource.NewRoleBindingList,
		enterFn: showRole,
	}
	m["rbac.authorization.k8s.io/v1/roles"] = resCmd{
		listFn:  resource.NewRoleList,
		enterFn: showRBAC,
	}
}

func extRes(m map[string]resCmd) {
	m["apiextensions.k8s.io/v1/customresourcedefinitions"] = resCmd{
		listFn:  resource.NewCustomResourceDefinitionList,
		enterFn: showCRD,
	}
}

func netRes(m map[string]resCmd) {
	m["networking.k8s.io/v1/networkpolicies"] = resCmd{
		gvr:    "apiextensions.k8s.io/NetworkPolicies",
		listFn: resource.NewNetworkPolicyList,
	}
	m["networking.k8s.io/v1beta1/ingresses"] = resCmd{
		listFn: resource.NewIngressList,
	}
}

func batchRes(m map[string]resCmd) {
	m["batch/v1/cronjobs"] = resCmd{
		viewFn: newCronJobView,
		listFn: resource.NewCronJobList,
	}
	m["batch/v1/jobs"] = resCmd{
		viewFn: newJobView,
		listFn: resource.NewJobList,
	}
}

func policyRes(m map[string]resCmd) {
	m["policy/v1beta1/poddisruptionbudgets"] = resCmd{
		viewFn:    newResourceView,
		listFn:    resource.NewPDBList,
		colorerFn: pdbColorer,
	}
}

func hpaRes(m map[string]resCmd) {
	m["autoscaling/v1/horizontalpodautoscalers"] = resCmd{
		listFn: resource.NewHorizontalPodAutoscalerV1List,
	}
}
