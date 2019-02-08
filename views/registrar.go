package views

import (
	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	"github.com/gdamore/tcell"
)

type (
	viewFn    func(string, *appView, resource.List, colorerFn) resourceViewer
	listFn    func(string) resource.List
	colorerFn func(string, *resource.RowEvent) tcell.Color

	resCmd struct {
		title     string
		api       string
		viewFn    viewFn
		listFn    listFn
		colorerFn colorerFn
	}
)

var cmdMap = map[string]resCmd{
	"cm": {
		title:     "Config Maps",
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewConfigMapList,
		colorerFn: defaultColorer,
	},
	"cr": {
		title:     "Cluster Roles",
		api:       "rbac.authorization.k8s.io",
		viewFn:    newResourceView,
		listFn:    resource.NewClusterRoleList,
		colorerFn: defaultColorer,
	},
	"crb": {
		title:     "Cluster Role Bindings",
		api:       "rbac.authorization.k8s.io",
		viewFn:    newResourceView,
		listFn:    resource.NewClusterRoleBindingList,
		colorerFn: defaultColorer,
	},
	"crd": {
		title:     "Custom Resource Definitions",
		api:       "apiextensions.k8s.io",
		viewFn:    newResourceView,
		listFn:    resource.NewCRDList,
		colorerFn: defaultColorer,
	},
	"cjo": {
		title:     "CronJobs",
		api:       "batch",
		viewFn:    newResourceView,
		listFn:    resource.NewCronJobList,
		colorerFn: defaultColorer,
	},
	"ctx": {
		title:     "Contexts",
		api:       "core",
		viewFn:    newContextView,
		listFn:    resource.NewContextList,
		colorerFn: ctxColorer,
	},
	"ds": {
		title:     "DaemonSets",
		api:       "core",
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
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewEndpointsList,
		colorerFn: defaultColorer,
	},
	"ev": {
		title:     "Events",
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewEventList,
		colorerFn: evColorer,
	},
	"hpa": {
		title:     "Horizontal Pod Autoscalers",
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
		viewFn:    newResourceView,
		listFn:    resource.NewJobList,
		colorerFn: defaultColorer,
	},
	"no": {
		title:     "Nodes",
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewNodeList,
		colorerFn: nsColorer,
	},
	"ns": {
		title:     "Namespaces",
		api:       "core",
		viewFn:    newNamespaceView,
		listFn:    resource.NewNamespaceList,
		colorerFn: nsColorer,
	},
	"po": {
		title:     "Pods",
		api:       "core",
		viewFn:    newPodView,
		listFn:    resource.NewPodList,
		colorerFn: podColorer,
	},
	"pv": {
		title:     "Persistent Volumes",
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewPVList,
		colorerFn: pvColorer,
	},
	"pvc": {
		title:     "Persistent Volume Claims",
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewPVCList,
		colorerFn: pvcColorer,
	},
	"rb": {
		title:     "Role Bindings",
		api:       "rbac.authorization.k8s.io",
		viewFn:    newResourceView,
		listFn:    resource.NewRoleBindingList,
		colorerFn: defaultColorer,
	},
	"rc": {
		title:     "ReplicationController",
		api:       "v1",
		viewFn:    newResourceView,
		listFn:    resource.NewReplicationControllerList,
		colorerFn: rsColorer,
	},
	"ro": {
		title:     "Roles",
		api:       "rbac.authorization.k8s.io",
		viewFn:    newResourceView,
		listFn:    resource.NewRoleList,
		colorerFn: defaultColorer,
	},
	"rs": {
		title:     "Replica Sets",
		api:       "apps",
		viewFn:    newResourceView,
		listFn:    resource.NewReplicaSetList,
		colorerFn: rsColorer,
	},
	"sa": {
		title:     "Service Accounts",
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewServiceAccountList,
		colorerFn: defaultColorer,
	},
	"sec": {
		title:     "Secrets",
		api:       "core",
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
		api:       "core",
		viewFn:    newResourceView,
		listFn:    resource.NewServiceList,
		colorerFn: defaultColorer,
	},
}

func helpCmds() map[string]resCmd {
	cmds := make(map[string]resCmd, len(cmdMap))
	for k, v := range cmdMap {
		cmds[k] = v
	}
	for k, v := range getCRDS() {
		cmds[k] = resCmd{title: v.Kind, api: v.Group}
	}
	return cmds
}

func getCRDS() map[string]k8s.ApiGroup {
	m := map[string]k8s.ApiGroup{}
	list := resource.NewCRDList(resource.AllNamespaces)
	ll, _ := list.Resource().List(resource.AllNamespaces)
	for _, l := range ll {
		ff := l.ExtFields()
		grp := k8s.ApiGroup{
			Version: ff["version"].(string),
			Group:   ff["group"].(string),
			Kind:    ff["kind"].(string),
		}
		if aa, ok := ff["aliases"].([]interface{}); ok {
			if n, ok := ff["plural"].(string); ok {
				grp.Plural = n
			}
			if n, ok := ff["singular"].(string); ok {
				grp.Singular = n
			}
			for _, a := range aa {
				m[a.(string)] = grp
			}
		} else if s, ok := ff["singular"].(string); ok {
			grp.Singular = s
			if p, ok := ff["plural"].(string); ok {
				grp.Plural = p
			}
			m[s] = grp
		} else if s, ok := ff["plural"].(string); ok {
			grp.Plural = s
			m[s] = grp
		}
	}
	return m
}
