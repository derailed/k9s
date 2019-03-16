package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
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

func helpCmds() map[string]resCmd {
	cmdMap := resourceViews()
	cmds := make(map[string]resCmd, len(cmdMap))
	for k, v := range cmdMap {
		cmds[k] = v
	}
	for k, v := range allCRDs() {
		cmds[k] = resCmd{title: v.Kind, api: v.Group}
	}

	return cmds
}

func allCRDs() map[string]k8s.APIGroup {
	m := map[string]k8s.APIGroup{}
	list := resource.NewCRDList(resource.AllNamespaces)
	ll, _ := list.Resource().List(resource.AllNamespaces)
	for _, l := range ll {
		ff := l.ExtFields()
		grp := k8s.APIGroup{
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
		"cron": {
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
		"job": {
			title:     "Jobs",
			api:       "batch",
			viewFn:    newJobView,
			listFn:    resource.NewJobList,
			colorerFn: defaultColorer,
		},
		"no": {
			title:     "Nodes",
			api:       "",
			viewFn:    newResourceView,
			listFn:    resource.NewNodeList,
			colorerFn: nsColorer,
		},
		"ns": {
			title:     "Namespaces",
			api:       "",
			viewFn:    newNamespaceView,
			listFn:    resource.NewNamespaceList,
			colorerFn: nsColorer,
		},
		"po": {
			title:     "Pods",
			api:       "",
			viewFn:    newPodView,
			listFn:    resource.NewPodList,
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
