package views

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	viewFn     func(ns string, app *appView, list resource.List) resourceViewer
	listFn     func(c resource.Connection, ns string, gvr k8s.GVR) resource.List
	enterFn    func(app *appView, ns, resource, selection string)
	decorateFn func(resource.TableData) resource.TableData

	aliases struct {
		plural   string
		singular string
	}

	resourcesCommand struct {
		aliases

		gvr k8s.GVR

		// groups are the possible groups a resource can be part of.
		// E.g., ingresses can appear on both networking and extensions,
		// this should be dynamically chosen based on the kubernetes cluster.
		groups     []string
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
		"dp":  "Deployment",
		"sec": "Secret",
		"jo":  "Jobs",

		"cr":  "ClusterRoles",
		"crb": "ClusterRoleBindings",
		"rb":  "RoleBindings",
		"ro":  "Roles",

		"np": "NetworkPolicies",

		"ctx": "Contexts",
		"usr": "Users",
		"grp": "Groups",
		"pf":  "PortForward",
		"be":  "Benchmark",
		"sd":  "ScreenDumps",
	},
}

func aliasCmds(c k8s.Connection, m map[string]*resourcesCommand) {
	resourceViews(c, m)
	if c != nil {
		allCRDs(c, m)
	}
}

func allCRDs(c k8s.Connection, m map[string]*resourcesCommand) {
	crds, _ := resource.NewCustomResourceDefinitionList(c, resource.AllNamespaces, k8s.GVR{}).
		Resource().
		List(resource.AllNamespaces)

	for _, crd := range crds {
		crdFields := crd.ExtFields()

		rsc := resourcesCommand{
			kind: crdFields["kind"].(string),
			gvr: k8s.GVR{
				Group:   crdFields["group"].(string),
				Version: crdFields["version"].(string),
			},
		}

		if p, ok := crdFields["plural"].(string); ok {
			rsc.plural = p
			m[p] = &rsc
			rsc.gvr.Resource = p
		}

		if s, ok := crdFields["singular"].(string); ok {
			rsc.singular = s
			m[s] = &rsc

			if rsc.gvr.Resource == "" {
				rsc.gvr.Resource = s
			}
		}

		if aa, ok := crdFields["aliases"].([]interface{}); ok {
			for _, a := range aa {
				m[a.(string)] = &rsc
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

func collect(commandList ...[]*resourcesCommand) (accum []*resourcesCommand) {
	for _, commands := range commandList {
		accum = append(accum, commands...)
	}
	return
}

func resourceViews(c k8s.Connection, cmdMap map[string]*resourcesCommand) {
	commands := collect(
		coreGroup(), storageGroup(), rbacGroup(), apiExtensionsGroup(),
		batchGroup(), appsGroup(), networkingGroup(), policyGroup(),
		customGroup(),
	)

	if c != nil {
		commands = append(commands, autoscalingGroup(c)...)
	}

	for _, rsc := range commands {
		cmdMap[strings.ToLower(rsc.kind)] = rsc
	}

	// Add default aliases
	// TODO: read aliases from a config file.
	for alias, kind := range DefaultAliasConfig.Aliases {
		if cmd, ok := cmdMap[strings.ToLower(kind)]; ok {
			addAlias(cmdMap, cmd, alias)
		}
	}

	if c != nil {
		discoverAliasesFromServer(c, cmdMap)
	}
}

func addAlias(cmdMap map[string]*resourcesCommand, cmd *resourcesCommand, alias string) {
	if alias == "" {
		return
	}

	alias = strings.ToLower(alias)
	if _, ok := cmdMap[alias]; !ok {
		cmdMap[alias] = cmd
	}
}

func addAPIResourceAliases(cmds map[string]*resourcesCommand, resource metav1.APIResource) {
	if strings.Contains(resource.Name, "/") {
		// Ignore resources that has slash, e.g.,
		// deployment/status, namespace/finalizers and etc.
		return
	}

	cmd, ok := cmds[strings.ToLower(resource.Kind)]
	if !ok {
		return
	}

	// Check if the group is the same we specified
	// or if it is one of the possible ones.
	if cmd.gvr.Group != resource.Group {
		found := false
		for _, group := range cmd.groups {
			if resource.Group == group {
				found = true
			}
		}

		if !found {
			return
		}
	}

	cmd.gvr.Version = resource.Version
	cmd.gvr.Group = resource.Group
	cmd.gvr.Resource = resource.Name

	addAlias(cmds, cmd, resource.Name)
	addAlias(cmds, cmd, resource.SingularName)
	for _, sn := range resource.ShortNames {
		addAlias(cmds, cmd, sn)
	}
}

func discoverAliasesFromServer(con k8s.Connection, cmds map[string]*resourcesCommand) {
	_, resourceLists, _ := con.DialOrDie().Discovery().ServerGroupsAndResources()
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			log.Fatal().Msgf("Invalid group version '%s' from server: %s", resourceList.GroupVersion, err.Error())
		}

		for _, resource := range resourceList.APIResources {
			resource.Version = gv.Version
			resource.Group = gv.Group
			addAPIResourceAliases(cmds, resource)
		}
	}
}

func storageGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind: "StorageClass",
			gvr: k8s.GVR{
				Group: "storage.k8s.io",
			},
			viewFn: newResourceView,
			listFn: resource.NewStorageClassList,
		},
	}
}

func coreGroup() []*resourcesCommand {
	return []*resourcesCommand{
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

func customGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind:      "Contexts",
			viewFn:    newContextView,
			listFn:    resource.NewContextList,
			colorerFn: ctxColorer,
		},
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

func rbacGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind: "ClusterRole",
			gvr: k8s.GVR{
				Group: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewClusterRoleList,
			enterFn: showRBAC,
		},
		{
			kind: "ClusterRoleBinding",
			gvr: k8s.GVR{
				Group: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewClusterRoleBindingList,
			enterFn: showClusterRole,
		},
		{
			kind: "RoleBinding",
			gvr: k8s.GVR{
				Group: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewRoleBindingList,
			enterFn: showRole,
		},
		{
			kind: "Role",
			gvr: k8s.GVR{
				Group: "rbac.authorization.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewRoleList,
			enterFn: showRBAC,
		},
	}
}

func apiExtensionsGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind: "CustomResourceDefinition",
			gvr: k8s.GVR{
				Group: "apiextensions.k8s.io",
			},
			viewFn:  newResourceView,
			listFn:  resource.NewCustomResourceDefinitionList,
			enterFn: showCRD,
		},
	}
}

func batchGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind: "CronJob",
			gvr: k8s.GVR{
				Group: "batch",
			},
			viewFn: newCronJobView,
			listFn: resource.NewCronJobList,
		},
		{
			kind: "Job",
			gvr: k8s.GVR{
				Group: "batch",
			},
			viewFn: newJobView,
			listFn: resource.NewJobList,
		},
	}
}

func appsGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind: "Deployment",
			gvr: k8s.GVR{
				Group: "apps",
			},
			viewFn:    newDeployView,
			listFn:    resource.NewDeploymentList,
			colorerFn: dpColorer,
		},
		{
			kind: "ReplicaSet",
			gvr: k8s.GVR{
				Group: "apps",
			},
			viewFn:    newReplicaSetView,
			listFn:    resource.NewReplicaSetList,
			colorerFn: rsColorer,
		},
		{
			kind: "DaemonSet",
			gvr: k8s.GVR{
				Group: "apps",
			},
			viewFn:    newDaemonSetView,
			listFn:    resource.NewDaemonSetList,
			colorerFn: dpColorer,
		},
		{
			kind: "StatefulSet",
			gvr: k8s.GVR{
				Group: "apps",
			},
			viewFn:    newStatefulSetView,
			listFn:    resource.NewStatefulSetList,
			colorerFn: stsColorer,
		},
	}

}

func networkingGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind: "Ingress",
			gvr: k8s.GVR{
				Group: "networking.k8s.io",
			},
			groups: []string{"networking.k8s.io", "extensions"},
			viewFn: newResourceView,
			listFn: resource.NewIngressList,
		},
		{
			kind: "NetworkPolicy",
			gvr: k8s.GVR{
				Group: "networking.k8s.io",
			},
			viewFn: newResourceView,
			listFn: resource.NewNetworkPolicyList,
		},
	}
}

func policyGroup() []*resourcesCommand {
	return []*resourcesCommand{
		{
			kind: "PodDisruptionBudget",
			gvr: k8s.GVR{
				Group: "policy",
			},
			viewFn:    newResourceView,
			listFn:    resource.NewPDBList,
			colorerFn: pdbColorer,
		},
	}
}

func autoscalingGroup(c k8s.Connection) []*resourcesCommand {
	rev, ok, err := c.SupportsRes("autoscaling", []string{"v1", "v2beta1", "v2beta2"})
	if err != nil {
		log.Error().Err(err).Msg("Checking HPA")
		return nil
	}
	if !ok {
		log.Error().Msg("HPA are not supported on this cluster")
		return nil
	}

	hpa := &resourcesCommand{
		kind: "HorizontalPodAutoscaler",
		gvr: k8s.GVR{
			Group: "autoscaling",
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

	return []*resourcesCommand{hpa}
}
