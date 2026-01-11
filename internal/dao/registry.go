// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	crdCat   = "crd"
	k9sCat   = "k9s"
	helmCat  = "helm"
	scaleCat = "scale"
)

var stdGroups = sets.New[string](
	"apps/v1",
	"autoscaling/v1",
	"autoscaling/v2",
	"autoscaling/v2beta1",
	"autoscaling/v2beta2",
	"batch/v1",
	"batch/v1beta1",
	"extensions/v1beta1",
	"policy/v1beta1",
	"policy/v1",
	"v1",
)

var scalableRes = sets.New(client.DpGVR, client.StsGVR, client.RsGVR, client.RcGVR)

// ResourceMetas represents a collection of resource metadata.
type ResourceMetas map[*client.GVR]*metav1.APIResource

func (m ResourceMetas) clear() {
	for k := range m {
		delete(m, k)
	}
}

// MetaAccess tracks resources metadata.
var MetaAccess = NewMeta()

// Meta represents available resource metas.
type Meta struct {
	resMetas ResourceMetas
	mx       sync.RWMutex
}

// NewMeta returns a resource meta.
func NewMeta() *Meta {
	return &Meta{resMetas: make(ResourceMetas)}
}

func (m *Meta) Lookup(cmd string) *client.GVR {
	m.mx.RLock()
	defer m.mx.RUnlock()
	for gvr, meta := range m.resMetas {
		if slices.Contains(meta.ShortNames, cmd) {
			return gvr
		}
		if meta.Name == cmd || meta.SingularName == cmd || meta.Kind == cmd {
			return gvr
		}
	}

	return client.NoGVR
}

// RegisterMeta registers a new resource meta object.
func (m *Meta) RegisterMeta(gvr string, res *metav1.APIResource) {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.resMetas[client.NewGVR(gvr)] = res
}

// AllGVRs returns all sorted cluster resources.
func (m *Meta) AllGVRs() client.GVRs {
	m.mx.RLock()
	defer m.mx.RUnlock()
	kk := slices.Collect(maps.Keys(m.resMetas))

	return client.GVRs(kk)
}

// GVK2GVR convert gvk to gvr
func (m *Meta) GVK2GVR(gv schema.GroupVersion, kind string) (*client.GVR, bool, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	for gvr, meta := range m.resMetas {
		if gv.Group == meta.Group && gv.Version == meta.Version && kind == meta.Kind {
			return gvr, meta.Namespaced, true
		}
	}

	return client.NoGVR, false, false
}

// IsNamespaced checks if a given resource is namespaced.
func (m *Meta) IsNamespaced(gvr *client.GVR) (bool, error) {
	res, err := m.MetaFor(gvr)
	if err != nil {
		return false, err
	}

	return res.Namespaced, nil
}

// MetaFor returns a resource metadata for a given gvr.
func (m *Meta) MetaFor(gvr *client.GVR) (*metav1.APIResource, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if meta, ok := m.resMetas[gvr]; ok {
		return meta, nil
	}

	return new(metav1.APIResource), fmt.Errorf("no resource meta defined for\n %q", gvr)
}

// IsCRD checks if resource represents a CRD
func IsCRD(r *metav1.APIResource) bool {
	return slices.Contains(r.Categories, crdCat)
}

// IsK8sMeta checks for non resource meta.
func IsK8sMeta(m *metav1.APIResource) bool {
	return !slices.ContainsFunc(m.Categories, func(category string) bool {
		return category == k9sCat || category == helmCat
	})
}

// IsK9sMeta checks for non resource meta.
func IsK9sMeta(m *metav1.APIResource) bool {
	return slices.Contains(m.Categories, k9sCat)
}

// IsScalable check if the resource can be scaled
func IsScalable(m *metav1.APIResource) bool {
	return slices.Contains(m.Categories, scaleCat)
}

// LoadResources hydrates server preferred+CRDs resource metadata.
func (m *Meta) LoadResources(f Factory) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.resMetas.clear()
	if err := loadPreferred(f, m.resMetas); err != nil {
		return err
	}
	loadNonResource(m.resMetas)

	// We've actually loaded all the CRDs in loadPreferred, and we're now adding
	// some additional CRD properties on top of that.
	loadCRDs(f, m.resMetas)

	return nil
}

// BOZO!! Need countermeasures for direct commands!
func loadNonResource(m ResourceMetas) {
	loadK9s(m)
	loadRBAC(m)
	loadHelm(m)
}

func loadK9s(m ResourceMetas) {
	m[client.WkGVR] = &metav1.APIResource{
		Name:         "workloads",
		Kind:         "Workload",
		SingularName: "workload",
		Namespaced:   true,
		ShortNames:   []string{"wk"},
		Categories:   []string{k9sCat},
	}
	m[client.PuGVR] = &metav1.APIResource{
		Name:         "pulses",
		Kind:         "Pulse",
		SingularName: "pulse",
		ShortNames:   []string{"hz", "pu"},
		Categories:   []string{k9sCat},
	}
	m[client.DirGVR] = &metav1.APIResource{
		Name:         "dirs",
		Kind:         "Dir",
		SingularName: "dir",
		Categories:   []string{k9sCat},
	}
	m[client.XGVR] = &metav1.APIResource{
		Name:         "xrays",
		Kind:         "XRays",
		SingularName: "xray",
		Categories:   []string{k9sCat},
	}
	m[client.RefGVR] = &metav1.APIResource{
		Name:         "references",
		Kind:         "References",
		SingularName: "reference",
		Verbs:        []string{},
		Categories:   []string{k9sCat},
	}
	m[client.AliGVR] = &metav1.APIResource{
		Name:         "aliases",
		Kind:         "Aliases",
		SingularName: "alias",
		Verbs:        []string{},
		Categories:   []string{k9sCat},
	}
	m[client.CtGVR] = &metav1.APIResource{
		Name:         client.CtGVR.String(),
		Kind:         "Contexts",
		SingularName: "context",
		ShortNames:   []string{"ctx"},
		Verbs:        []string{},
		Categories:   []string{k9sCat},
	}
	m[client.SdGVR] = &metav1.APIResource{
		Name:         "screendumps",
		Kind:         "ScreenDumps",
		SingularName: "screendump",
		ShortNames:   []string{"sd"},
		Verbs:        []string{"delete"},
		Categories:   []string{k9sCat},
	}
	m[client.BeGVR] = &metav1.APIResource{
		Name:         "benchmarks",
		Kind:         "Benchmarks",
		SingularName: "benchmark",
		ShortNames:   []string{"be"},
		Verbs:        []string{"delete"},
		Categories:   []string{k9sCat},
	}
	m[client.PfGVR] = &metav1.APIResource{
		Name:         "portforwards",
		Namespaced:   true,
		Kind:         "PortForwards",
		SingularName: "portforward",
		ShortNames:   []string{"pf"},
		Verbs:        []string{"delete"},
		Categories:   []string{k9sCat},
	}
	m[client.CoGVR] = &metav1.APIResource{
		Name:         "containers",
		Kind:         "Containers",
		SingularName: "container",
		Verbs:        []string{},
		Categories:   []string{k9sCat},
	}
	m[client.ScnGVR] = &metav1.APIResource{
		Name:         "scans",
		Kind:         "Scans",
		SingularName: "scan",
		Verbs:        []string{},
		Categories:   []string{k9sCat},
	}
}

func loadHelm(m ResourceMetas) {
	m[client.HmGVR] = &metav1.APIResource{
		Name:       "helm",
		Kind:       "Helm",
		Namespaced: true,
		Verbs:      []string{"delete"},
		Categories: []string{helmCat},
	}
	m[client.HmhGVR] = &metav1.APIResource{
		Name:       "history",
		Kind:       "History",
		Namespaced: true,
		Verbs:      []string{"delete"},
		Categories: []string{helmCat},
	}
}

func loadRBAC(m ResourceMetas) {
	m[client.RbacGVR] = &metav1.APIResource{
		Name:       "rbacs",
		Kind:       "Rules",
		Categories: []string{k9sCat},
	}
	m[client.PolGVR] = &metav1.APIResource{
		Name:       "policies",
		Kind:       "Rules",
		Namespaced: true,
		Categories: []string{k9sCat},
	}
	m[client.UsrGVR] = &metav1.APIResource{
		Name:       "users",
		Kind:       "User",
		Categories: []string{k9sCat},
	}
	m[client.GrpGVR] = &metav1.APIResource{
		Name:       "groups",
		Kind:       "Group",
		Categories: []string{k9sCat},
	}
}

func loadPreferred(f Factory, m ResourceMetas) error {
	if f == nil || f.Client() == nil || !f.Client().ConnectionOK() {
		slog.Error("Load cluster resources - No API server connection")
		return nil
	}

	dial, err := f.Client().CachedDiscovery()
	if err != nil {
		return err
	}
	rr, err := dial.ServerPreferredResources()
	if err != nil {
		slog.Debug("Failed to load preferred resources", slogs.Error, err)
	}
	for _, r := range rr {
		for i := range r.APIResources {
			res := r.APIResources[i]
			gvr := client.FromGVAndR(r.GroupVersion, res.Name)
			if isDeprecated(gvr) {
				continue
			}
			res.Group, res.Version = gvr.G(), gvr.V()
			if res.SingularName == "" {
				res.SingularName = strings.ToLower(res.Kind)
			}
			if !isStandardGroup(r.GroupVersion) {
				res.Categories = append(res.Categories, crdCat)
			}
			if isScalable(gvr) {
				res.Categories = append(res.Categories, scaleCat)
			}
			m[gvr] = &res
		}
	}

	return nil
}

func isStandardGroup(gv string) bool {
	return stdGroups.Has(gv) || strings.Contains(gv, ".k8s.io")
}

func isScalable(gvr *client.GVR) bool {
	return scalableRes.Has(gvr)
}

var deprecatedGVRs = sets.New(
	client.NewGVR("v1/events"),
	client.NewGVR("extensions/v1beta1/ingresses"),
)

func isDeprecated(gvr *client.GVR) bool {
	return deprecatedGVRs.Has(gvr) || gvr.V() == ""
}

// loadCRDs Wait for the cache to synced and then add some additional properties to CRD.
func loadCRDs(f Factory, m ResourceMetas) {
	if f == nil || f.Client() == nil || !f.Client().ConnectionOK() {
		return
	}

	oo, err := f.List(client.CrdGVR, client.ClusterScope, true, labels.Everything())
	if err != nil {
		slog.Warn("CRDs load Fail", slogs.Error, err)
		return
	}

	for _, o := range oo {
		var crd apiext.CustomResourceDefinition
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crd)
		if err != nil {
			slog.Error("CRD conversion failed", slogs.Error, err)
			continue
		}
		for gvr, version := range client.NewGVRFromCRD(&crd) {
			if meta, ok := m[gvr]; ok && version.Subresources != nil && version.Subresources.Scale != nil {
				if !slices.Contains(meta.Categories, scaleCat) {
					meta.Categories = append(meta.Categories, scaleCat)
					m[gvr] = meta
				}
			}
		}
	}
}
