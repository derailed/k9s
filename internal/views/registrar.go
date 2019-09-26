package views

import (
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

		kind       string
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

func aliasCmds(c k8s.Connection, m map[string]*resCmd) {
	resourceViews(c, m)
	if c != nil {
		allCRDs(c, m)
	}
}

func allCRDs(c k8s.Connection, m map[string]*resCmd) {
	crds, _ := resource.NewCustomResourceDefinitionList(c, resource.AllNamespaces).
		Resource().
		List(resource.AllNamespaces)

	for _, crd := range crds {
		ff := crd.ExtFields()

		grp := k8s.APIGroup{
			GKV: k8s.GKV{
				Group:   ff["group"].(string),
				Kind:    ff["kind"].(string),
				Version: ff["version"].(string),
			},
		}

		res := resCmd{
			kind: grp.Kind,
			crdCmd: crdCmd{
				api:     grp.Group,
				version: grp.Version,
			},
		}
		if p, ok := ff["plural"].(string); ok {
			res.plural = p
			m[p] = &res
		}

		if s, ok := ff["singular"].(string); ok {
			res.singular = s
			m[s] = &res
		}

		if aa, ok := ff["aliases"].([]interface{}); ok {
			for _, a := range aa {
				m[a.(string)] = &res
			}
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

func collect(commandList ...[]*resCmd) (accum []*resCmd) {
	for _, commands := range commandList {
		accum = append(accum, commands...)
	}
	return
}

func resourceViews(c k8s.Connection, cmdMap map[string]*resCmd) {
	commands := collect(
		primRes(), coreRes(), stateRes(), rbacRes(), apiExtRes(),
		batchRes(), appsRes(), extRes(), v1beta1Res(), custRes(),
	)

	if c != nil {
		commands = append(commands, hpaRes(c)...)
	}

	for _, rsc := range commands {
		cmdMap[strings.ToLower(rsc.kind)] = rsc
	}

	// Add default aliases
	// TODO: read aliases from a config file.
	for rsc, alias := range DefaultAliasConfig.Aliases {
		if cmd, ok := cmdMap[strings.ToLower(rsc)]; ok {
			addAlias(cmdMap, cmd, alias)
		}
	}

	if c != nil {
		discoverAliasesFromServer(c, cmdMap)
	}
}

func addAlias(cmdMap map[string]*resCmd, cmd *resCmd, alias string) {
	if alias == "" {
		return
	}

	alias = strings.ToLower(alias)
	if _, ok := cmdMap[alias]; !ok {
		cmdMap[alias] = cmd
	}
}

func addAPIResourceAliases(cmds map[string]*resCmd, resource metav1.APIResource) {
	if strings.Contains(resource.Name, "/") {
		// Ignore resources that has slash, e.g.,
		// deployment/status, namespace/finalizers and etc.
		return
	}

	if cmd, ok := cmds[strings.ToLower(resource.Kind)]; ok {
		addAlias(cmds, cmd, resource.Name)
		addAlias(cmds, cmd, resource.SingularName)
		for _, sn := range resource.ShortNames {
			addAlias(cmds, cmd, sn)
		}
	}
}

func discoverAliasesFromServer(con k8s.Connection, cmds map[string]*resCmd) {
	_, resourceLists, _ := con.DialOrDie().Discovery().ServerGroupsAndResources()
	for _, resourceList := range resourceLists {
		for _, resource := range resourceList.APIResources {
			addAPIResourceAliases(cmds, resource)
		}
	}
}

func stateRes() []*resCmd {
	return []*resCmd{
		{
			kind:   "ConfigMap",
			viewFn: newResourceView,
			listFn: resource.NewConfigMapList,
		},
		{
			kind:      "PersistentVolume",
			viewFn:    newResourceView,
			listFn:    resource.NewPersistentVolumeList,
			colorerFn: pvColorer,
		},
		{
			kind:      "PersistentVolumeClaim",
			viewFn:    newResourceView,
			listFn:    resource.NewPersistentVolumeClaimList,
			colorerFn: pvcColorer,
		},
		{
			kind:   "Secret",
			viewFn: newSecretView,
			listFn: resource.NewSecretList,
		},
		{
			kind: "StorageClass",
			crdCmd: crdCmd{
				api: "storage.k8s.io",
			},
			viewFn: newResourceView,
			listFn: resource.NewStorageClassList,
		},
	}
}

func primRes() []*resCmd {
	return []*resCmd{
		{
			kind:      "Node",
			viewFn:    newNodeView,
			listFn:    resource.NewNodeList,
			colorerFn: nsColorer,
		},
		{
			kind:      "Namespace",
			viewFn:    newNamespaceView,
			listFn:    resource.NewNamespaceList,
			colorerFn: nsColorer,
		},
		{
			kind:      "Pod",
			viewFn:    newPodView,
			listFn:    resource.NewPodList,
			colorerFn: podColorer,
		},
		{
			kind:    "ServiceAccount",
			viewFn:  newResourceView,
			listFn:  resource.NewServiceAccountList,
			enterFn: showSAPolicy,
		},
		{
			kind:   "Service",
			viewFn: newSvcView,
			listFn: resource.NewServiceList,
		},
	}
}

func coreRes() []*resCmd {
	return []*resCmd{
		{
			kind:      "Contexts",
			viewFn:    newContextView,
			listFn:    resource.NewContextList,
			colorerFn: ctxColorer,
		},
		{
			kind:      "DaemonSet",
			viewFn:    newDaemonSetView,
			listFn:    resource.NewDaemonSetList,
			colorerFn: dpColorer,
		},
		{
			kind:   "EndPoints",
			viewFn: newResourceView,
			listFn: resource.NewEndpointsList,
		},
		{
			kind:      "Event",
			viewFn:    newResourceView,
			listFn:    resource.NewEventList,
			colorerFn: evColorer,
		},
		{
			kind:      "ReplicationController",
			viewFn:    newScalableResourceView,
			listFn:    resource.NewReplicationControllerList,
			colorerFn: rsColorer,
		},
	}
}

func custRes() []*resCmd {
	return []*resCmd{
		{
			kind:   "Users",
			viewFn: newSubjectView,
		},
		{
			kind:   "Groups",
			viewFn: newSubjectView,
		},
		{
			kind:   "PortForward",
			viewFn: newForwardView,
		},
		{
			kind:   "Benchmark",
			viewFn: newBenchView,
		},
		{
			kind:   "ScreenDumps",
			viewFn: newDumpView,
		},
	}

}

func rbacRes() []*resCmd {
	return []*resCmd{
		{
			kind: "ClusterRole",
			crdCmd: crdCmd{
				api: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewClusterRoleList,
			enterFn: showRBAC,
		},
		{
			kind: "ClusterRoleBinding",
			crdCmd: crdCmd{
				api: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewClusterRoleBindingList,
			enterFn: showClusterRole,
		},
		{
			kind: "RoleBinding",
			crdCmd: crdCmd{
				api: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewRoleBindingList,
			enterFn: showRole,
		},
		{
			kind: "Role",
			crdCmd: crdCmd{
				api: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewRoleList,
			enterFn: showRBAC,
		},
	}
}

func apiExtRes() []*resCmd {
	return []*resCmd{
		{
			kind: "CustomResourceDefinition",
			crdCmd: crdCmd{
				api: "apiextensions.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewCustomResourceDefinitionList,
			enterFn: showCRD,
		},
		{
			kind: "NetworkPolicy",
			crdCmd: crdCmd{
				api: "apiextensions.k8s.io",
			},
			viewFn: newResourceView,
			listFn: resource.NewNetworkPolicyList,
		},
	}
}

func batchRes() []*resCmd {
	return []*resCmd{
		{
			kind: "CronJob",
			crdCmd: crdCmd{
				api: "batch",
			},
			viewFn: newCronJobView,
			listFn: resource.NewCronJobList,
		},
		{
			kind: "Job",
			crdCmd: crdCmd{
				api: "batch",
			},
			viewFn: newJobView,
			listFn: resource.NewJobList,
		},
	}
}

func appsRes() []*resCmd {
	return []*resCmd{
		{
			kind: "Deployment",
			crdCmd: crdCmd{
				api: "apps",
			},
			viewFn:    newDeployView,
			listFn:    resource.NewDeploymentList,
			colorerFn: dpColorer,
		},
		{
			kind: "ReplicaSet",
			crdCmd: crdCmd{
				api: "apps",
			},
			viewFn:    newReplicaSetView,
			listFn:    resource.NewReplicaSetList,
			colorerFn: rsColorer,
		},
		{
			kind: "StatefulSet",
			crdCmd: crdCmd{
				api: "apps",
			},
			viewFn:    newStatefulSetView,
			listFn:    resource.NewStatefulSetList,
			colorerFn: stsColorer,
		},
	}

}

func extRes() []*resCmd {
	return []*resCmd{
		{
			kind: "Ingress",
			crdCmd: crdCmd{
				api: "extensions",
			},
			viewFn: newResourceView,
			listFn: resource.NewIngressList,
		},
	}
}

func v1beta1Res() []*resCmd {
	return []*resCmd{
		{
			kind: "PodDisruptionBudget",
			crdCmd: crdCmd{
				api: "v1.beta1",
			},
			viewFn:    newResourceView,
			listFn:    resource.NewPDBList,
			colorerFn: pdbColorer,
		},
	}
}

func hpaRes(c k8s.Connection) []*resCmd {
	rev, ok, err := c.SupportsRes("autoscaling", []string{"v1", "v2beta1", "v2beta2"})
	if err != nil {
		log.Error().Err(err).Msg("Checking HPA")
		return nil
	}
	if !ok {
		log.Error().Msg("HPA are not supported on this cluster")
		return nil
	}

	hpa := &resCmd{
		kind: "HorizontalPodAutoscaler",
		crdCmd: crdCmd{
			api: "autoscaling",
		},
		viewFn: newResourceView,
	}

	switch rev {
	case "v1":
		hpa.listFn = resource.NewHorizontalPodAutoscalerV1List
	case "v2beta1":
		hpa.listFn = resource.NewHorizontalPodAutoscalerV2Beta1List
	case "v2beta2":
		hpa.listFn = resource.NewHorizontalPodAutoscalerList
	default:
		log.Panic().Msgf("K9s unsupported HPA version. Exiting!")
	}

	return []*resCmd{hpa}
}
