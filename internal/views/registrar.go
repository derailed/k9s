package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	viewFn     func(ns string, app *appView, list resource.List) resourceViewer
	listFn     func(c resource.Connection, ns string) resource.List
	colorerFn  func(ns string, evt *resource.RowEvent) tcell.Color
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

		title      string
		viewFn     viewFn
		listFn     listFn
		enterFn    enterFn
		colorerFn  colorerFn
		decorateFn decorateFn
	}
)

func aliasCmds(c k8s.Connection, m map[string]resCmd) {
	resourceViews(c, m)
	if c != nil {
		allCRDs(c, m)
	}
}

func allCRDs(c k8s.Connection, m map[string]resCmd) {
	crds, _ := resource.NewCustomResourceDefinitionList(c, resource.AllNamespaces).
		Resource().
		List(resource.AllNamespaces)

	for _, crd := range crds {
		ff := crd.ExtFields()

		grp := k8s.APIGroup{
			Group:   ff["group"].(string),
			Kind:    ff["kind"].(string),
			Version: ff["version"].(string),
		}

		res := resCmd{
			title: grp.Kind,
			crdCmd: crdCmd{
				api:     grp.Group,
				version: grp.Version,
			},
		}
		if p, ok := ff["plural"].(string); ok {
			res.plural = p
			m[p] = res
		}

		if s, ok := ff["singular"].(string); ok {
			res.singular = s
			m[s] = res
		}

		if aa, ok := ff["aliases"].([]interface{}); ok {
			for _, a := range aa {
				m[a.(string)] = res
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

func showClusterRole(app *appView, ns, resource, selection string) {
	crb, err := app.conn().DialOrDie().Rbac().ClusterRoleBindings().Get(selection, metav1.GetOptions{})
	if err != nil {
		app.flash().errf("Unable to retrieve clusterrolebindings for %s", selection)
		return
	}
	app.inject(newRBACView(app, ns, crb.RoleRef.Name, clusterRole))
}

func showRole(app *appView, _, resource, selection string) {
	ns, n := namespaced(selection)
	rb, err := app.conn().DialOrDie().Rbac().RoleBindings(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		app.flash().errf("Unable to retrieve rolebindings for %s", selection)
		return
	}
	app.inject(newRBACView(app, ns, fqn(ns, rb.RoleRef.Name), role))
}

func showSAPolicy(app *appView, _, _, selection string) {
	_, n := namespaced(selection)
	app.inject(newPolicyView(app, mapFuSubject("ServiceAccount"), n))
}

func resourceViews(c k8s.Connection, m map[string]resCmd) {
	coreRes(m)
	rbacRes(m)
	apiExtRes(m)
	batchRes(m)
	appsRes(m)
	extRes(m)
	v1beta1Res(m)
	custRes(m)

	if c != nil {
		hpaRes(c, m)
	}
}

func coreRes(m map[string]resCmd) {
	m["cm"] = resCmd{
		title: "ConfigMaps",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newResourceView,
		listFn: resource.NewConfigMapList,
	}
	m["ctx"] = resCmd{
		title: "Contexts",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newContextView,
		listFn:    resource.NewContextList,
		colorerFn: ctxColorer,
	}
	m["ds"] = resCmd{
		title: "DaemonSets",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newDaemonSetView,
		listFn:    resource.NewDaemonSetList,
		colorerFn: dpColorer,
	}
	m["ep"] = resCmd{
		title: "EndPoints",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newResourceView,
		listFn: resource.NewEndpointsList,
	}
	m["ev"] = resCmd{
		title: "Events",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newResourceView,
		listFn:    resource.NewEventList,
		colorerFn: evColorer,
	}
	m["no"] = resCmd{
		title: "Nodes",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newNodeView,
		listFn:    resource.NewNodeList,
		colorerFn: nsColorer,
	}
	m["ns"] = resCmd{
		title: "Namespaces",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newNamespaceView,
		listFn:    resource.NewNamespaceList,
		colorerFn: nsColorer,
	}
	m["po"] = resCmd{
		title: "Pods",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newPodView,
		listFn:    resource.NewPodList,
		colorerFn: podColorer,
	}
	m["pv"] = resCmd{
		title: "PersistentVolumes",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newResourceView,
		listFn:    resource.NewPersistentVolumeList,
		colorerFn: pvColorer,
	}
	m["pvc"] = resCmd{
		title: "PersistentVolumeClaims",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newResourceView,
		listFn:    resource.NewPersistentVolumeClaimList,
		colorerFn: pvcColorer,
	}
	m["rc"] = resCmd{
		title: "ReplicationControllers",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:    newResourceView,
		listFn:    resource.NewReplicationControllerList,
		colorerFn: rsColorer,
	}
	m["sa"] = resCmd{
		title: "ServiceAccounts",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn:  newResourceView,
		listFn:  resource.NewServiceAccountList,
		enterFn: showSAPolicy,
	}
	m["sec"] = resCmd{
		title: "Secrets",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newSecretView,
		listFn: resource.NewSecretList,
	}
	m["svc"] = resCmd{
		title: "Services",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newSvcView,
		listFn: resource.NewServiceList,
	}
}

func custRes(m map[string]resCmd) {
	m["usr"] = resCmd{
		title: "Users",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newSubjectView,
	}
	m["grp"] = resCmd{
		title: "Groups",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newSubjectView,
	}
	m["pf"] = resCmd{
		title: "PortForward",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newForwardView,
	}
	m["be"] = resCmd{
		title: "Benchmark",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newBenchView,
	}
	m["sd"] = resCmd{
		title: "ScreenDumps",
		crdCmd: crdCmd{
			api: "",
		},
		viewFn: newDumpView,
	}
}

func rbacRes(m map[string]resCmd) {
	m["cr"] = resCmd{
		title: "ClusterRoles",
		crdCmd: crdCmd{
			api: "rbac.authorization.k8s.io",
		},
		viewFn:  newResourceView,
		listFn:  resource.NewClusterRoleList,
		enterFn: showRBAC,
	}
	m["crb"] = resCmd{
		title: "ClusterRoleBindings",
		crdCmd: crdCmd{
			api: "rbac.authorization.k8s.io",
		},
		viewFn:  newResourceView,
		listFn:  resource.NewClusterRoleBindingList,
		enterFn: showClusterRole,
	}

	m["rb"] = resCmd{
		title: "RoleBindings",
		crdCmd: crdCmd{
			api: "rbac.authorization.k8s.io",
		},
		viewFn:  newResourceView,
		listFn:  resource.NewRoleBindingList,
		enterFn: showRole,
	}
	m["ro"] = resCmd{
		title: "Roles",
		crdCmd: crdCmd{
			api: "rbac.authorization.k8s.io",
		},
		viewFn:  newResourceView,
		listFn:  resource.NewRoleList,
		enterFn: showRBAC,
	}
}

func apiExtRes(m map[string]resCmd) {
	m["crd"] = resCmd{
		title: "CustomResourceDefinitions",
		crdCmd: crdCmd{
			api: "apiextensions.k8s.io",
		},
		viewFn: newResourceView,
		listFn: resource.NewCustomResourceDefinitionList,
	}
}

func batchRes(m map[string]resCmd) {
	m["cj"] = resCmd{
		title: "CronJobs",
		crdCmd: crdCmd{
			api: "batch",
		},
		viewFn: newCronJobView,
		listFn: resource.NewCronJobList,
	}
	m["jo"] = resCmd{
		title: "Jobs",
		crdCmd: crdCmd{
			api: "batch",
		},
		viewFn: newJobView,
		listFn: resource.NewJobList,
	}
}

func appsRes(m map[string]resCmd) {
	m["dp"] = resCmd{
		title: "Deployments",
		crdCmd: crdCmd{
			api: "apps",
		},
		viewFn:    newDeployView,
		listFn:    resource.NewDeploymentList,
		colorerFn: dpColorer,
	}
	m["rs"] = resCmd{
		title: "ReplicaSets",
		crdCmd: crdCmd{
			api: "apps",
		},
		viewFn:    newReplicaSetView,
		listFn:    resource.NewReplicaSetList,
		colorerFn: rsColorer,
	}
	m["sts"] = resCmd{
		title: "StatefulSets",
		crdCmd: crdCmd{
			api: "apps",
		},
		viewFn:    newStatefulSetView,
		listFn:    resource.NewStatefulSetList,
		colorerFn: stsColorer,
	}
}

func extRes(m map[string]resCmd) {
	m["ing"] = resCmd{
		title: "Ingress",
		crdCmd: crdCmd{
			api: "extensions",
		},
		viewFn: newResourceView,
		listFn: resource.NewIngressList,
	}
}

func v1beta1Res(m map[string]resCmd) {
	m["pdb"] = resCmd{
		title: "PodDisruptionBudgets",
		crdCmd: crdCmd{
			api: "v1.beta1",
		},
		viewFn:    newResourceView,
		listFn:    resource.NewPDBList,
		colorerFn: pdbColorer,
	}
}

func hpaRes(c k8s.Connection, cmds map[string]resCmd) {
	rev, ok, err := c.SupportsRes("autoscaling", []string{"v1", "v2beta1", "v2beta2"})
	if err != nil {
		log.Error().Err(err).Msg("Checking HPA")
		return
	}
	if !ok {
		log.Error().Msg("HPA are not supported on this cluster")
		return
	}

	switch rev {
	case "v1":
		cmds["hpa"] = resCmd{
			title: "HorizontalPodAutoscalers",
			crdCmd: crdCmd{
				api: "autoscaling",
			},
			viewFn: newResourceView,
			listFn: resource.NewHorizontalPodAutoscalerV1List,
		}
	case "v2beta1":
		cmds["hpa"] = resCmd{
			title: "HorizontalPodAutoscalers",
			crdCmd: crdCmd{
				api: "autoscaling",
			},
			viewFn: newResourceView,
			listFn: resource.NewHorizontalPodAutoscalerV2Beta1List,
		}
	case "v2beta2":
		cmds["hpa"] = resCmd{
			title: "HorizontalPodAutoscalers",
			crdCmd: crdCmd{
				api: "autoscaling",
			},
			viewFn: newResourceView,
			listFn: resource.NewHorizontalPodAutoscalerList,
		}
	default:
		log.Panic().Msgf("K9s unsupported HPA version. Exiting!")
	}
}
