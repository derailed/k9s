package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type (
	viewFn    func(ns string, app *appView, list resource.List, colorer colorerFn) resourceViewer
	listFn    func(c resource.Connection, ns string) resource.List
	listMxFn  func(c resource.Connection, mx resource.MetricsServer, ns string) resource.List
	colorerFn func(ns string, evt *resource.RowEvent) tcell.Color
	enterFn   func(app *appView, ns, resource, selection string)

	resCmd struct {
		title     string
		api       string
		viewFn    viewFn
		listFn    listFn
		listMxFn  listMxFn
		enterFn   enterFn
		colorerFn colorerFn
	}
)

func helpCmds(c k8s.Connection) map[string]resCmd {
	cmdMap := resourceViews()
	cmds := make(map[string]resCmd, len(cmdMap))
	for k, v := range cmdMap {
		cmds[k] = v
	}
	for k, v := range allCRDs(c) {
		cmds[k] = resCmd{title: v.Kind, api: v.Group}
	}

	return cmds
}

func allCRDs(c k8s.Connection) map[string]k8s.APIGroup {
	m := map[string]k8s.APIGroup{}

	crds, _ := resource.
		NewCRDList(c, resource.AllNamespaces).
		Resource().
		List(resource.AllNamespaces)

	for _, crd := range crds {
		ff := crd.ExtFields()

		grp := k8s.APIGroup{
			Group:   ff["group"].(string),
			Kind:    ff["kind"].(string),
			Version: ff["version"].(string),
		}

		if p, ok := ff["plural"].(string); ok {
			grp.Plural = p
			m[p] = grp
		}

		if s, ok := ff["singular"].(string); ok {
			grp.Singular = s
			m[s] = grp
		}

		if aa, ok := ff["aliases"].([]interface{}); ok {
			for _, a := range aa {
				m[a.(string)] = grp
			}
		}
	}

	return m
}

func showRBAC(app *appView, ns, resource, selection string) {
	log.Debug().Msgf("Entered FN on `%s`--%s:%s", ns, resource, selection)
	kind := clusterRole
	if resource == "role" {
		kind = role
	}
	app.command.pushCmd("policies")
	app.inject(newRBACView(app, ns, selection, kind))
}

func resourceViews() map[string]resCmd {
	return map[string]resCmd{
		"cm": {
			title:     "ConfigMaps",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewConfigMapList,
			colorerFn: defaultColorer,
		},
		"cr": {
			title:     "ClusterRoles",
			api:       "rbac.authorization.k8s.io",
			viewFn:    newResourceView,
			listFn:    resource.NewClusterRoleList,
			enterFn:   showRBAC,
			colorerFn: defaultColorer,
		},
		"crb": {
			title:     "ClusterRoleBindings",
			api:       "rbac.authorization.k8s.io",
			viewFn:    newResourceView,
			listFn:    resource.NewClusterRoleBindingList,
			colorerFn: defaultColorer,
		},
		"crd": {
			title:     "CustomResourceDefinitions",
			api:       "apiextensions.k8s.io",
			viewFn:    newResourceView,
			listFn:    resource.NewCRDList,
			colorerFn: defaultColorer,
		},
		"cj": {
			title:     "CronJobs",
			api:       "batch",
			viewFn:    newCronJobView,
			listFn:    resource.NewCronJobList,
			colorerFn: defaultColorer,
		},
		"ctx": {
			title:     "Contexts",
			api:       "",
			viewFn:    newContextView,
			listFn:    resource.NewContextList,
			colorerFn: ctxColorer,
		},
		"ds": {
			title:     "DaemonSets",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewDaemonSetList,
			colorerFn: dpColorer,
		},
		"dp": {
			title:     "Deployments",
			api:       "apps",
			viewFn:    newResourceView,
			listFn:    resource.NewDeploymentList,
			colorerFn: dpColorer,
		},
		"ep": {
			title:     "EndPoints",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewEndpointsList,
			colorerFn: defaultColorer,
		},
		"ev": {
			title:     "Events",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewEventList,
			colorerFn: evColorer,
		},
		"hpa": {
			title:     "HorizontalPodAutoscalers",
			api:       "autoscaling",
			viewFn:    newResourceView,
			listFn:    resource.NewHPAList,
			colorerFn: defaultColorer,
		},
		"ing": {
			title:     "Ingress",
			api:       "extensions",
			viewFn:    newResourceView,
			listFn:    resource.NewIngressList,
			colorerFn: defaultColorer,
		},
		"jo": {
			title:     "Jobs",
			api:       "batch",
			viewFn:    newJobView,
			listFn:    resource.NewJobList,
			colorerFn: defaultColorer,
		},
		"no": {
			title:     "Nodes",
			api:       "",
			viewFn:    newNodeView,
			listMxFn:  resource.NewNodeList,
			colorerFn: nsColorer,
		},
		"ns": {
			title:     "Namespaces",
			api:       "",
			viewFn:    newNamespaceView,
			listFn:    resource.NewNamespaceList,
			colorerFn: nsColorer,
		},
		"pdb": {
			title:     "PodDisruptionBudgets",
			api:       "v1.beta1",
			viewFn:    newResourceView,
			listFn:    resource.NewPDBList,
			colorerFn: pdbColorer,
		},
		"po": {
			title:     "Pods",
			api:       "",
			viewFn:    newPodView,
			listMxFn:  resource.NewPodList,
			colorerFn: podColorer,
		},
		"pv": {
			title:     "PersistentVolumes",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewPVList,
			colorerFn: pvColorer,
		},
		"pvc": {
			title:     "PersistentVolumeClaims",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewPVCList,
			colorerFn: pvcColorer,
		},
		"rb": {
			title:     "RoleBindings",
			api:       "rbac.authorization.k8s.io",
			viewFn:    newResourceView,
			listFn:    resource.NewRoleBindingList,
			colorerFn: defaultColorer,
		},
		"rc": {
			title:     "ReplicationControllers",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewReplicationControllerList,
			colorerFn: rsColorer,
		},
		"ro": {
			title:     "Roles",
			api:       "rbac.authorization.k8s.io",
			viewFn:    newResourceView,
			listFn:    resource.NewRoleList,
			enterFn:   showRBAC,
			colorerFn: defaultColorer,
		},
		"rs": {
			title:     "ReplicaSets",
			api:       "apps",
			viewFn:    newResourceView,
			listFn:    resource.NewReplicaSetList,
			colorerFn: rsColorer,
		},
		"sa": {
			title:     "ServiceAccounts",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewServiceAccountList,
			colorerFn: defaultColorer,
		},
		"sec": {
			title:     "Secrets",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewSecretList,
			colorerFn: defaultColorer,
		},
		"sts": {
			title:     "StatefulSets",
			api:       "apps",
			viewFn:    newResourceView,
			listFn:    resource.NewStatefulSetList,
			colorerFn: stsColorer,
		},
		"svc": {
			title:     "Services",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewServiceList,
			colorerFn: defaultColorer,
		},
	}
}
