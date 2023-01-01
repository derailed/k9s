package dao

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// CRD identifies a CRD.
const CRD = "crd"

// MetaAccess tracks resources metadata.
var MetaAccess = NewMeta()

var stdGroups = []string{
	"admissionregistration.k8s.io/v1",
	"admissionregistration.k8s.io/v1beta1",
	"apiextensions.k8s.io/v1",
	"apiextensions.k8s.io/v1beta1",
	"apiregistration.k8s.io/v1",
	"apiregistration.k8s.io/v1beta1",
	"apps/v1",
	"authentication.k8s.io/v1",
	"authentication.k8s.io/v1beta1",
	"authorization.k8s.io/v1",
	"authorization.k8s.io/v1beta1",
	"autoscaling/v1",
	"autoscaling/v2beta1",
	"autoscaling/v2beta2",
	"batch/v1",
	"batch/v1beta1",
	"certificates.k8s.io/v1",
	"certificates.k8s.io/v1beta1",
	"coordination.k8s.io/v1",
	"coordination.k8s.io/v1beta1",
	"discovery.k8s.io/v1beta1",
	"dynatrace.com/v1alpha1",
	"events.k8s.io/v1",
	"extensions/v1beta1",
	"flowcontrol.apiserver.k8s.io/v1beta1",
	"metrics.k8s.io/v1beta1",
	"networking.k8s.io/v1",
	"networking.k8s.io/v1beta1",
	"node.k8s.io/v1",
	"node.k8s.io/v1beta1",
	"policy/v1beta1",
	"rbac.authorization.k8s.io/v1",
	"rbac.authorization.k8s.io/v1beta1",
	"scheduling.k8s.io/v1",
	"scheduling.k8s.io/v1beta1",
	"storage.k8s.io/v1",
	"storage.k8s.io/v1beta1",
	"v1",
}

// Meta represents available resource metas.
type Meta struct {
	resMetas ResourceMetas
	mx       sync.RWMutex
}

// NewMeta returns a resource meta.
func NewMeta() *Meta {
	return &Meta{resMetas: make(ResourceMetas)}
}

// AccessorFor returns a client accessor for a resource if registered.
// Otherwise it returns a generic accessor.
// Customize here for non resource types or types with metrics or logs.
func AccessorFor(f Factory, gvr client.GVR) (Accessor, error) {
	m := Accessors{
		client.NewGVR("contexts"):               &Context{},
		client.NewGVR("containers"):             &Container{},
		client.NewGVR("screendumps"):            &ScreenDump{},
		client.NewGVR("benchmarks"):             &Benchmark{},
		client.NewGVR("portforwards"):           &PortForward{},
		client.NewGVR("v1/services"):            &Service{},
		client.NewGVR("v1/pods"):                &Pod{},
		client.NewGVR("v1/nodes"):               &Node{},
		client.NewGVR("apps/v1/deployments"):    &Deployment{},
		client.NewGVR("apps/v1/daemonsets"):     &DaemonSet{},
		client.NewGVR("apps/v1/statefulsets"):   &StatefulSet{},
		client.NewGVR("batch/v1/cronjobs"):      &CronJob{},
		client.NewGVR("batch/v1beta1/cronjobs"): &CronJob{},
		client.NewGVR("batch/v1/jobs"):          &Job{},
		client.NewGVR("v1/namespaces"):          &Namespace{},
		// BOZO!! Revamp with latest...
		// client.NewGVR("openfaas"):               &OpenFaas{},
		client.NewGVR("popeye"):    &Popeye{},
		client.NewGVR("sanitizer"): &Popeye{},
		client.NewGVR("helm"):      &Helm{},
		client.NewGVR("dir"):       &Dir{},
	}

	r, ok := m[gvr]
	if !ok {
		r = &Generic{}
		log.Debug().Msgf("No DAO registry entry for %q. Using generics!", gvr)
	}
	r.Init(f, gvr)

	return r, nil
}

// RegisterMeta registers a new resource meta object.
func (m *Meta) RegisterMeta(gvr string, res metav1.APIResource) {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.resMetas[client.NewGVR(gvr)] = res
}

// AllGVRs returns all cluster resources.
func (m *Meta) AllGVRs() client.GVRs {
	m.mx.RLock()
	defer m.mx.RUnlock()

	kk := make(client.GVRs, 0, len(m.resMetas))
	for k := range m.resMetas {
		kk = append(kk, k)
	}
	sort.Sort(kk)

	return kk
}

// IsCRD checks if resource represents a CRD
func IsCRD(r v1.APIResource) bool {
	for _, c := range r.Categories {
		if c == CRD {
			return true
		}
	}
	return false
}

// MetaFor returns a resource metadata for a given gvr.
func (m *Meta) MetaFor(gvr client.GVR) (metav1.APIResource, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	meta, ok := m.resMetas[gvr]
	if !ok {
		return metav1.APIResource{}, fmt.Errorf("no resource meta defined for %q", gvr)
	}
	return meta, nil
}

// IsK8sMeta checks for non resource meta.
func IsK8sMeta(m metav1.APIResource) bool {
	for _, c := range m.Categories {
		if c == "k9s" || c == "helm" || c == "faas" {
			return false
		}
	}

	return true
}

// IsK9sMeta checks for non resource meta.
func IsK9sMeta(m metav1.APIResource) bool {
	for _, c := range m.Categories {
		if c == "k9s" {
			return true
		}
	}

	return false
}

// LoadResources hydrates server preferred+CRDs resource metadata.
func (m *Meta) LoadResources(f Factory) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.resMetas = make(ResourceMetas, 100)
	if err := loadPreferred(f, m.resMetas); err != nil {
		return err
	}
	loadNonResource(m.resMetas)
	loadCRDs(f, m.resMetas)

	return nil
}

// BOZO!! Need countermeasures for direct commands!
func loadNonResource(m ResourceMetas) {
	loadK9s(m)
	loadRBAC(m)
	loadHelm(m)
	// BOZO!! Revamp with latest...
	// if IsOpenFaasEnabled() {
	// 	loadOpenFaas(m)
	// }
}

func loadK9s(m ResourceMetas) {
	m[client.NewGVR("pulses")] = metav1.APIResource{
		Name:         "pulses",
		Kind:         "Pulse",
		SingularName: "pulses",
		ShortNames:   []string{"hz", "pu"},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("dir")] = metav1.APIResource{
		Name:         "dir",
		Kind:         "Dir",
		SingularName: "dir",
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("xrays")] = metav1.APIResource{
		Name:         "xray",
		Kind:         "XRays",
		SingularName: "xray",
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("references")] = metav1.APIResource{
		Name:         "references",
		Kind:         "References",
		SingularName: "reference",
		Verbs:        []string{},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("aliases")] = metav1.APIResource{
		Name:         "aliases",
		Kind:         "Aliases",
		SingularName: "alias",
		Verbs:        []string{},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("popeye")] = metav1.APIResource{
		Name:         "popeye",
		Kind:         "Popeye",
		SingularName: "popeye",
		Namespaced:   true,
		Verbs:        []string{},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("sanitizer")] = metav1.APIResource{
		Name:         "sanitizer",
		Kind:         "Sanitizer",
		SingularName: "sanitizer",
		Verbs:        []string{},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("contexts")] = metav1.APIResource{
		Name:         "contexts",
		Kind:         "Contexts",
		SingularName: "context",
		ShortNames:   []string{"ctx"},
		Verbs:        []string{},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("screendumps")] = metav1.APIResource{
		Name:         "screendumps",
		Kind:         "ScreenDumps",
		SingularName: "screendump",
		ShortNames:   []string{"sd"},
		Verbs:        []string{"delete"},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("benchmarks")] = metav1.APIResource{
		Name:         "benchmarks",
		Kind:         "Benchmarks",
		SingularName: "benchmark",
		ShortNames:   []string{"be"},
		Verbs:        []string{"delete"},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("portforwards")] = metav1.APIResource{
		Name:         "portforwards",
		Namespaced:   true,
		Kind:         "PortForwards",
		SingularName: "portforward",
		ShortNames:   []string{"pf"},
		Verbs:        []string{"delete"},
		Categories:   []string{"k9s"},
	}
	m[client.NewGVR("containers")] = metav1.APIResource{
		Name:         "containers",
		Kind:         "Containers",
		SingularName: "container",
		Verbs:        []string{},
		Categories:   []string{"k9s"},
	}
}

func loadHelm(m ResourceMetas) {
	m[client.NewGVR("helm")] = metav1.APIResource{
		Name:       "helm",
		Kind:       "Helm",
		Namespaced: true,
		Verbs:      []string{"delete"},
		Categories: []string{"helm"},
	}
}

// BOZO!! revamp with latest...
// func loadOpenFaas(m ResourceMetas) {
// 	m[client.NewGVR("openfaas")] = metav1.APIResource{
// 		Name:       "openfaas",
// 		Kind:       "OpenFaaS",
// 		ShortNames: []string{"ofaas", "ofa"},
// 		Namespaced: true,
// 		Verbs:      []string{"delete"},
// 		Categories: []string{"faas"},
// 	}
// }

func loadRBAC(m ResourceMetas) {
	m[client.NewGVR("rbac")] = metav1.APIResource{
		Name:       "rbacs",
		Kind:       "Rules",
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("policy")] = metav1.APIResource{
		Name:       "policies",
		Kind:       "Rules",
		Namespaced: true,
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("users")] = metav1.APIResource{
		Name:       "users",
		Kind:       "User",
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("groups")] = metav1.APIResource{
		Name:       "groups",
		Kind:       "Group",
		Categories: []string{"k9s"},
	}
}

func loadPreferred(f Factory, m ResourceMetas) error {
	if !f.Client().ConnectionOK() {
		log.Error().Msgf("Load cluster resources - No API server connection")
		return nil
	}

	dial, err := f.Client().CachedDiscovery()
	if err != nil {
		return err
	}
	rr, err := dial.ServerPreferredResources()
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to load preferred resources")
	}
	for _, r := range rr {
		for _, res := range r.APIResources {
			gvr := client.FromGVAndR(r.GroupVersion, res.Name)
			if isDeprecated(gvr) {
				continue
			}
			res.Group, res.Version = gvr.G(), gvr.V()
			if res.SingularName == "" {
				res.SingularName = strings.ToLower(res.Kind)
			}
			if !isStandardGroup(res.Group) {
				res.Categories = append(res.Categories, CRD)
			}
			m[gvr] = res
		}
	}

	return nil
}

func isStandardGroup(r string) bool {
	for _, res := range stdGroups {
		if strings.Index(res, r) == 0 {
			return true
		}
	}

	return false
}

var deprecatedGVRs = map[client.GVR]struct{}{
	client.NewGVR("extensions/v1beta1/ingresses"): {},
}

func isDeprecated(gvr client.GVR) bool {
	_, ok := deprecatedGVRs[gvr]
	return ok
}

func loadCRDs(f Factory, m ResourceMetas) {
	if !f.Client().ConnectionOK() {
		return
	}
	const crdGVR = "apiextensions.k8s.io/v1/customresourcedefinitions"
	oo, err := f.List(crdGVR, client.ClusterScope, false, labels.Everything())
	if err != nil {
		log.Warn().Err(err).Msgf("Fail CRDs load")
		return
	}

	for _, o := range oo {
		meta, errs := extractMeta(o)
		if len(errs) > 0 {
			log.Error().Err(errs[0]).Msgf("Fail to extract CRD meta (%d) errors", len(errs))
			continue
		}
		meta.Categories = append(meta.Categories, CRD)
		gvr := client.NewGVRFromMeta(meta)
		m[gvr] = meta
	}
}

func extractMeta(o runtime.Object) (metav1.APIResource, []error) {
	var (
		m    metav1.APIResource
		errs []error
	)

	crd, ok := o.(*unstructured.Unstructured)
	if !ok {
		return m, append(errs, fmt.Errorf("Expected Unstructured, but got %T", o))
	}

	var spec map[string]interface{}
	spec, errs = extractMap(crd.Object, "spec", errs)

	var meta map[string]interface{}
	meta, errs = extractMap(crd.Object, "metadata", errs)
	m.Name, errs = extractStr(meta, "name", errs)

	m.Group, errs = extractStr(spec, "group", errs)
	versions, errs := extractSlice(spec, "versions", errs)
	if len(versions) > 0 {
		m.Version = versions[0]
	}

	var scope string
	scope, errs = extractStr(spec, "scope", errs)

	m.Namespaced = isNamespaced(scope)

	var names map[string]interface{}
	names, errs = extractMap(spec, "names", errs)
	m.Kind, errs = extractStr(names, "kind", errs)
	m.SingularName, errs = extractStr(names, "singular", errs)
	m.Name, errs = extractStr(names, "plural", errs)
	m.ShortNames, errs = extractSlice(names, "shortNames", errs)

	return m, errs
}

func isNamespaced(scope string) bool {
	return scope == "Namespaced"
}

func extractSlice(m map[string]interface{}, n string, errs []error) ([]string, []error) {
	if m[n] == nil {
		return nil, errs
	}

	s, ok := m[n].([]string)
	if ok {
		return s, errs
	}

	ii, ok := m[n].([]interface{})
	if !ok {
		return s, append(errs, fmt.Errorf("failed to extract slice %s -- %#v", n, m))
	}

	ss := make([]string, 0, len(ii))
	for _, name := range ii {
		switch o := name.(type) {
		case string:
			ss = append(ss, o)
		case map[string]interface{}:
			s, ok := o["name"].(string)
			if ok {
				ss = append(ss, s)
			} else {
				errs = append(errs, fmt.Errorf("unable to find key %q in map", n))
			}
		default:
			errs = append(errs, fmt.Errorf("unknown field type %t for key %q", o, n))
		}
	}

	return ss, errs
}

func extractStr(m map[string]interface{}, n string, errs []error) (string, []error) {
	s, ok := m[n].(string)
	if !ok {
		return s, append(errs, fmt.Errorf("failed to extract string %s", n))
	}
	return s, errs
}

func extractMap(m map[string]interface{}, n string, errs []error) (map[string]interface{}, []error) {
	v, ok := m[n].(map[string]interface{})
	if !ok {
		return v, append(errs, fmt.Errorf("failed to extract field %s", n))
	}
	return v, errs
}
