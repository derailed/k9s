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

type (
	viewFn     func(title, gvr string, list resource.List) ResourceViewer
	listFn     func(c resource.Connection, ns string) resource.List
	enterFn    func(app *App, ns, resource, selection string)
	decorateFn func(resource.TableData) resource.TableData

	viewerCapability struct {
		viewFn     viewFn
		listFn     listFn
		enterFn    enterFn
		colorerFn  ui.ColorerFunc
		decorateFn decorateFn
	}

	viewer struct {
		viewerCapability

		gvr        string
		kind       string
		namespaced bool
		verbs      metav1.Verbs
	}

	viewers map[string]viewer
)

func listFunc(l resource.List) viewFn {
	return func(title, gvr string, list resource.List) ResourceViewer {
		return NewResource(title, gvr, l)
	}
}

var aliases = config.NewAliases()

func allCRDs(c k8s.Connection, vv viewers) {
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

		vv[gvrs] = viewer{
			gvr:  gvrs,
			kind: meta.Kind,
			viewerCapability: viewerCapability{
				viewFn:    listFunc(resource.NewCustomList(c, meta.Namespaced, "", gvrs)),
				colorerFn: ui.DefaultColorer,
			},
		}
	}
}

func showRBAC(app *App, ns, resource, selection string) {
	kind := ClusterRole
	if resource == "role" {
		kind = Role
	}
	app.inject(NewRbac(app, ns, selection, kind))
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
	app.inject(NewRbac(app, ns, crb.RoleRef.Name, ClusterRole))
}

func showRole(app *App, _, resource, selection string) {
	ns, n := namespaced(selection)
	rb, err := app.Conn().DialOrDie().RbacV1().RoleBindings(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve rolebindings for %s", selection)
		return
	}
	app.inject(NewRbac(app, ns, fqn(ns, rb.RoleRef.Name), Role))
}

func showSAPolicy(app *App, _, _, selection string) {
	_, n := namespaced(selection)
	app.inject(NewPolicy(app, mapFuSubject("ServiceAccount"), n))
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

func resourceViews(c k8s.Connection, m viewers) {
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
		viewerCapability: viewerCapability{
			viewFn:    NewNode,
			listFn:    resource.NewNodeList,
			colorerFn: nsColorer,
		},
	}
	vv["v1/namespaces"] = viewer{
		viewerCapability: viewerCapability{
			viewFn:    NewNamespace,
			listFn:    resource.NewNamespaceList,
			colorerFn: nsColorer,
		},
	}
	vv["v1/pods"] = viewer{
		viewerCapability: viewerCapability{
			viewFn:    NewPod,
			listFn:    resource.NewPodList,
			colorerFn: podColorer,
		},
	}
	vv["v1/serviceaccounts"] = viewer{
		viewerCapability: viewerCapability{
			listFn:  resource.NewServiceAccountList,
			enterFn: showSAPolicy,
		},
	}
	vv["v1/services"] = viewer{
		viewerCapability: viewerCapability{
			viewFn: NewService,
			listFn: resource.NewServiceList,
		},
	}
	vv["v1/configmaps"] = viewer{
		viewerCapability: viewerCapability{
			listFn: resource.NewConfigMapList,
		},
	}
	vv["v1/persistentvolumes"] = viewer{
		viewerCapability: viewerCapability{
			listFn:    resource.NewPersistentVolumeList,
			colorerFn: pvColorer,
		},
	}
	vv["v1/persistentvolumeclaims"] = viewer{
		viewerCapability: viewerCapability{
			listFn:    resource.NewPersistentVolumeClaimList,
			colorerFn: pvcColorer,
		},
	}
	vv["v1/secrets"] = viewer{
		viewerCapability: viewerCapability{
			viewFn: NewSecret,
			listFn: resource.NewSecretList,
		},
	}
	vv["v1/endpoints"] = viewer{
		viewerCapability: viewerCapability{
			listFn: resource.NewEndpointsList,
		},
	}
	vv["v1/events"] = viewer{
		viewerCapability: viewerCapability{
			listFn:    resource.NewEventList,
			colorerFn: evColorer,
		},
	}
	vv["v1/replicationcontrollers"] = viewer{
		viewerCapability: viewerCapability{
			viewFn:    NewScalableResource,
			listFn:    resource.NewReplicationControllerList,
			colorerFn: rsColorer,
		},
	}
}

func miscRes(vv viewers) {
	vv["storage.k8s.io/v1/storageclasses"] = viewer{
		viewerCapability: viewerCapability{
			listFn: resource.NewStorageClassList,
		},
	}
	vv["contexts"] = viewer{
		gvr:  "contexts",
		kind: "Contexts",
		viewerCapability: viewerCapability{
			viewFn:    NewContext,
			listFn:    resource.NewContextList,
			colorerFn: ctxColorer,
		},
	}
	vv["users"] = viewer{
		gvr: "users",
		viewerCapability: viewerCapability{
			viewFn: NewSubject,
		},
	}
	vv["groups"] = viewer{
		gvr: "groups",
		viewerCapability: viewerCapability{
			viewFn: NewSubject,
		},
	}
	vv["portforwards"] = viewer{
		gvr: "portforwards",
		viewerCapability: viewerCapability{
			viewFn: NewPortForward,
		},
	}
	vv["benchmarks"] = viewer{
		gvr: "benchmarks",
		viewerCapability: viewerCapability{
			viewFn: NewBench,
		},
	}
	vv["screendumps"] = viewer{
		gvr: "screendumps",
		viewerCapability: viewerCapability{
			viewFn: NewScreenDump,
		},
	}
}

func appsRes(vv viewers) {
	vv["apps/v1/deployments"] = viewer{
		viewerCapability: viewerCapability{
			viewFn:    NewDeploy,
			listFn:    resource.NewDeploymentList,
			colorerFn: dpColorer,
		},
	}
	vv["apps/v1/replicasets"] = viewer{
		viewerCapability: viewerCapability{
			viewFn:    NewReplicaSet,
			listFn:    resource.NewReplicaSetList,
			colorerFn: rsColorer,
		},
	}
	vv["apps/v1/statefulsets"] = viewer{
		viewerCapability: viewerCapability{
			viewFn:    NewStatefulSet,
			listFn:    resource.NewStatefulSetList,
			colorerFn: stsColorer,
		},
	}
	vv["apps/v1/daemonsets"] = viewer{
		viewerCapability: viewerCapability{
			viewFn:    NewDaemonSet,
			listFn:    resource.NewDaemonSetList,
			colorerFn: dpColorer,
		},
	}
}

func authRes(vv viewers) {
	vv["rbac.authorization.k8s.io/v1/clusterroles"] = viewer{
		viewerCapability: viewerCapability{
			listFn:  resource.NewClusterRoleList,
			enterFn: showRBAC,
		},
	}
	vv["rbac.authorization.k8s.io/v1/clusterrolebindings"] = viewer{
		viewerCapability: viewerCapability{
			listFn:  resource.NewClusterRoleBindingList,
			enterFn: showClusterRole,
		},
	}
	vv["rbac.authorization.k8s.io/v1/rolebindings"] = viewer{
		viewerCapability: viewerCapability{
			listFn:  resource.NewRoleBindingList,
			enterFn: showRole,
		},
	}
	vv["rbac.authorization.k8s.io/v1/roles"] = viewer{
		viewerCapability: viewerCapability{
			listFn:  resource.NewRoleList,
			enterFn: showRBAC,
		},
	}
}

func extRes(vv viewers) {
	vv["apiextensions.k8s.io/v1/customresourcedefinitions"] = viewer{
		viewerCapability: viewerCapability{
			listFn:  resource.NewCustomResourceDefinitionList,
			enterFn: showCRD,
		},
	}
	vv["apiextensions.k8s.io/v1beta1/customresourcedefinitions"] = viewer{
		viewerCapability: viewerCapability{
			listFn:  resource.NewCustomResourceDefinitionList,
			enterFn: showCRD,
		},
	}
}

func netRes(vv viewers) {
	vv["networking.k8s.io/v1/networkpolicies"] = viewer{
		viewerCapability: viewerCapability{
			listFn: resource.NewNetworkPolicyList,
		},
	}
	vv["extensions/v1beta1/ingresses"] = viewer{
		viewerCapability: viewerCapability{
			listFn: resource.NewIngressList,
		},
	}
}

func batchRes(vv viewers) {
	vv["batch/v1beta1/cronjobs"] = viewer{
		viewerCapability: viewerCapability{
			viewFn: NewCronJob,
			listFn: resource.NewCronJobList,
		},
	}
	vv["batch/v1/jobs"] = viewer{
		viewerCapability: viewerCapability{
			viewFn: NewJob,
			listFn: resource.NewJobList,
		},
	}
}

func policyRes(vv viewers) {
	vv["policy/v1beta1/poddisruptionbudgets"] = viewer{
		viewerCapability: viewerCapability{
			listFn:    resource.NewPDBList,
			colorerFn: pdbColorer,
		},
	}
}

func hpaRes(vv viewers) {
	vv["autoscaling/v1/horizontalpodautoscalers"] = viewer{
		viewerCapability: viewerCapability{
			listFn: resource.NewHorizontalPodAutoscalerV1List,
		},
	}
}
