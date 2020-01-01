package dao

import (
	"fmt"
	"sort"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceMetas represents a collection of resource metadata.
type ResourceMetas map[client.GVR]metav1.APIResource

// Accessors represents a collection of dao accessors.
type Accessors map[client.GVR]Accessor

var resMetas = ResourceMetas{}

// AccessorFor returns a client accessor for a resource if registered.
// Otherwise it returns a generic accessor.
// Customize here for non resource types or types with metrics or logs.
func AccessorFor(f Factory, gvr client.GVR) (Accessor, error) {
	m := Accessors{
		client.NewGVR("contexts"):                      &Context{},
		client.NewGVR("containers"):                    &Container{},
		client.NewGVR("screendumps"):                   &ScreenDump{},
		client.NewGVR("benchmarks"):                    &Benchmark{},
		client.NewGVR("portforwards"):                  &PortForward{},
		client.NewGVR("v1/services"):                   &Service{},
		client.NewGVR("v1/pods"):                       &Pod{},
		client.NewGVR("apps/v1/deployments"):           &Deployment{},
		client.NewGVR("apps/v1/daemonsets"):            &DaemonSet{},
		client.NewGVR("extensions/v1beta1/daemonsets"): &DaemonSet{},
		client.NewGVR("apps/v1/statefulsets"):          &StatefulSet{},
		client.NewGVR("batch/v1beta1/cronjobs"):        &CronJob{},
		client.NewGVR("batch/v1/jobs"):                 &Job{},
	}

	r, ok := m[gvr]
	if !ok {
		r = &Generic{}
		log.Warn().Msgf("No DAO registry entry for %q. Using factory!", gvr)
	}
	r.Init(f, gvr)

	return r, nil
}

// RegisterMeta registers a new resource meta object.
func RegisterMeta(gvr string, res metav1.APIResource) {
	resMetas[client.NewGVR(gvr)] = res
}

// AllGVRs returns all cluster resources.
func AllGVRs() client.GVRs {
	kk := make(client.GVRs, 0, len(resMetas))
	for k := range resMetas {
		kk = append(kk, k)
	}
	sort.Sort(kk)

	return kk
}

// MetaFor returns a resource metadata for a given gvr.
func MetaFor(gvr client.GVR) (metav1.APIResource, error) {
	m, ok := resMetas[gvr]
	if !ok {
		return metav1.APIResource{}, fmt.Errorf("no resource meta defined for %q", gvr)
	}
	return m, nil
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
func LoadResources(f Factory) error {
	resMetas = make(ResourceMetas, 100)
	if err := loadPreferred(f, resMetas); err != nil {
		return err
	}
	loadNonResource(resMetas)
	loadCRDs(f, resMetas)

	return nil
}

// BOZO!! Need contermeasure for direct commands!
func loadNonResource(m ResourceMetas) {
	m[client.NewGVR("aliases")] = metav1.APIResource{
		Name:       "aliases",
		Kind:       "Aliases",
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("contexts")] = metav1.APIResource{
		Name:       "contexts",
		Kind:       "Contexts",
		ShortNames: []string{"ctx"},
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("screendumps")] = metav1.APIResource{
		Name:       "screendumps",
		Kind:       "ScreenDumps",
		ShortNames: []string{"sd"},
		Verbs:      []string{"delete"},
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("benchmarks")] = metav1.APIResource{
		Name:       "benchmarks",
		Kind:       "Benchmarks",
		ShortNames: []string{"be"},
		Verbs:      []string{"delete"},
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("portforwards")] = metav1.APIResource{
		Name:       "portforwards",
		Namespaced: true,
		Kind:       "PortForwards",
		ShortNames: []string{"pf"},
		Verbs:      []string{"delete"},
		Categories: []string{"k9s"},
	}
	m[client.NewGVR("containers")] = metav1.APIResource{
		Name:       "containers",
		Kind:       "Containers",
		Categories: []string{"k9s"},
	}

	loadRBAC(m)
}

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
	discovery, err := f.Client().CachedDiscovery()
	if err != nil {
		return err
	}
	rr, err := discovery.ServerPreferredResources()
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to load preferred resources")
	}
	for _, r := range rr {
		for _, res := range r.APIResources {
			gvr := client.FromGVAndR(r.GroupVersion, res.Name)
			res.Group, res.Version = gvr.ToG(), gvr.ToV()
			m[gvr] = res
		}
	}

	return nil
}

func loadCRDs(f Factory, m ResourceMetas) {
	log.Debug().Msgf("Loading CRDs...")
	const crdGVR = "apiextensions.k8s.io/v1beta1/customresourcedefinitions"
	oo, err := f.List(crdGVR, "", true, labels.Everything())
	if err != nil {
		log.Warn().Err(err).Msgf("Fail CRDs load")
		return
	}
	log.Debug().Msgf(">>> CRDS count %d", len(oo))

	for _, o := range oo {
		meta, errs := extractMeta(o)
		if len(errs) > 0 {
			log.Error().Err(errs[0]).Msgf("Fail to extract CRD meta (%d) errors", len(errs))
			continue
		}
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
	m.Version, errs = extractStr(spec, "version", errs)

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

	ss := make([]string, len(ii))
	for i, name := range ii {
		ss[i], ok = name.(string)
		if !ok {
			return ss, append(errs, fmt.Errorf("expecting string shortnames"))
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
