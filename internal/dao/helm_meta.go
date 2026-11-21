// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render/helm"
	"github.com/derailed/k9s/internal/slogs"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/metadata"
)

const (
	helmOwnerSelector        = "owner=helm"
	helmPageSize       int64 = 500
	helmEnrichWorkers        = 8
	helmEnrichTimeout        = 30 * time.Second
	helmFailureBackoff       = time.Minute

	helmLabelName    = "name"
	helmLabelStatus  = "status"
	helmLabelVersion = "version"
)

// helmSummary holds the release details that are only available in the
// (large, compressed) release payload and not in the secret labels.
type helmSummary struct {
	chart, version, appVersion string
	deployed                   time.Time
}

type helmCacheEntry struct {
	ns  string
	sum helmSummary
}

// helmSummaries caches release summaries keyed by secret UID and resource
// version so each release revision payload is downloaded at most once.
type helmSummaries struct {
	mx       sync.Mutex
	done     map[string]helmCacheEntry
	inflight map[string]struct{}
	failed   map[string]time.Time
	sem      chan struct{}
}

func newHelmSummaries() *helmSummaries {
	return &helmSummaries{
		done:     make(map[string]helmCacheEntry),
		inflight: make(map[string]struct{}),
		failed:   make(map[string]time.Time),
		sem:      make(chan struct{}, helmEnrichWorkers),
	}
}

var helmCache = newHelmSummaries()

func helmCacheKey(uid types.UID, rv string) string {
	return string(uid) + "/" + rv
}

func (c *helmSummaries) get(key string) (helmSummary, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	e, ok := c.done[key]

	return e.sum, ok
}

// claim reserves a key for fetching. It returns false if the key is already
// being fetched or failed recently.
func (c *helmSummaries) claim(key string) bool {
	c.mx.Lock()
	defer c.mx.Unlock()

	if _, ok := c.inflight[key]; ok {
		return false
	}
	if t, ok := c.failed[key]; ok {
		if time.Since(t) < helmFailureBackoff {
			return false
		}
		delete(c.failed, key)
	}
	c.inflight[key] = struct{}{}

	return true
}

func (c *helmSummaries) finish(key, ns string, sum helmSummary, err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	delete(c.inflight, key)
	if err != nil {
		c.failed[key] = time.Now()
		return
	}
	c.done[key] = helmCacheEntry{ns: ns, sum: sum}
}

// prune drops entries in scope (ns == "" means every namespace) that are no
// longer part of the latest listing.
func (c *helmSummaries) prune(ns string, live map[string]struct{}) {
	c.mx.Lock()
	defer c.mx.Unlock()

	for k, e := range c.done {
		if _, ok := live[k]; ok {
			continue
		}
		if ns == "" || e.ns == ns {
			delete(c.done, k)
		}
	}
	for k, t := range c.failed {
		if _, ok := live[k]; !ok || time.Since(t) >= helmFailureBackoff {
			delete(c.failed, k)
		}
	}
}

// helmRev identifies the latest revision of a release.
type helmRev struct {
	ns, name, status string
	version          int
	uid              types.UID
	rv               string
	created          time.Time
}

func (r *helmRev) key() string {
	return helmCacheKey(r.uid, r.rv)
}

// latestHelmRevs reduces a list of helm secrets metadata to the latest
// revision of each release.
func latestHelmRevs(items []metav1.PartialObjectMetadata) []helmRev {
	latest := make(map[string]helmRev, len(items))
	for i := range items {
		it := &items[i]
		name := it.Labels[helmLabelName]
		v, err := strconv.Atoi(it.Labels[helmLabelVersion])
		if name == "" || err != nil {
			continue
		}
		k := it.Namespace + "/" + name
		if cur, ok := latest[k]; ok && cur.version >= v {
			continue
		}
		latest[k] = helmRev{
			ns:      it.Namespace,
			name:    name,
			status:  it.Labels[helmLabelStatus],
			version: v,
			uid:     it.UID,
			rv:      it.ResourceVersion,
			created: it.CreationTimestamp.Time,
		}
	}

	rr := make([]helmRev, 0, len(latest))
	for _, r := range latest {
		rr = append(rr, r)
	}

	return rr
}

// toRelease builds a lightweight release suitable for rendering. Chart is
// left nil until the summary has been loaded.
func (r *helmRev) toRelease(sum *helmSummary) *release.Release {
	rel := &release.Release{
		Name:      r.name,
		Namespace: r.ns,
		Version:   r.version,
		Info: &release.Info{
			Status:       release.Status(r.status),
			LastDeployed: helmtime.Time{Time: r.created},
		},
	}
	if sum != nil {
		rel.Info.LastDeployed = helmtime.Time{Time: sum.deployed}
		rel.Chart = &chart.Chart{Metadata: &chart.Metadata{
			Name:       sum.chart,
			Version:    sum.version,
			AppVersion: sum.appVersion,
		}}
	}

	return rel
}

func useHelmMetadata() bool {
	d := os.Getenv("HELM_DRIVER")
	return d == "" || d == "secret" || d == "secrets"
}

// listMeta lists the latest revision of every release using a metadata only
// request, avoiding the download of every release payload on each refresh.
func (h *HelmChart) listMeta(ctx context.Context, ns string) ([]runtime.Object, error) {
	if client.IsAllNamespace(ns) {
		ns = client.BlankNamespace
	}
	cfg, err := h.Client().RestConfig()
	if err != nil {
		return nil, err
	}
	mc, err := metadata.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	res := mc.Resource(client.SecGVR.GVR())

	var items []metav1.PartialObjectMetadata
	opts := metav1.ListOptions{LabelSelector: helmOwnerSelector, Limit: helmPageSize}
	for {
		var page *metav1.PartialObjectMetadataList
		if ns == client.BlankNamespace {
			page, err = res.List(ctx, opts)
		} else {
			page, err = res.Namespace(ns).List(ctx, opts)
		}
		if err != nil {
			return nil, fmt.Errorf("list helm releases metadata: %w", err)
		}
		items = append(items, page.Items...)
		if page.Continue == "" {
			break
		}
		opts.Continue = page.Continue
	}

	revs := latestHelmRevs(items)
	live := make(map[string]struct{}, len(revs))
	oo := make([]runtime.Object, 0, len(revs))
	var missing []helmRev
	for i := range revs {
		r := &revs[i]
		live[r.key()] = struct{}{}
		sum, ok := helmCache.get(r.key())
		if ok {
			oo = append(oo, helm.ReleaseRes{Release: r.toRelease(&sum)})
			continue
		}
		oo = append(oo, helm.ReleaseRes{Release: r.toRelease(nil)})
		missing = append(missing, *r)
	}
	helmCache.prune(ns, live)
	helmCache.enrich(h.Client().Config().Flags(), missing)

	return oo, nil
}

// enrich asynchronously loads the summaries for the given revisions.
func (c *helmSummaries) enrich(flags *genericclioptions.ConfigFlags, revs []helmRev) {
	for i := range revs {
		r := revs[i]
		if !c.claim(r.key()) {
			continue
		}
		go func() {
			c.sem <- struct{}{}
			defer func() { <-c.sem }()

			sum, err := fetchHelmSummary(flags, &r)
			if err != nil {
				slog.Warn("Unable to load helm release details",
					slogs.FQN, client.FQN(r.ns, r.name),
					slogs.Error, err,
				)
			}
			c.finish(r.key(), r.ns, sum, err)
		}()
	}
}

func fetchHelmSummary(flags *genericclioptions.ConfigFlags, r *helmRev) (helmSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), helmEnrichTimeout)
	defer cancel()

	type result struct {
		rel *release.Release
		err error
	}
	ch := make(chan result, 1)
	go func() {
		cfg, err := ensureHelmConfig(flags, r.ns)
		if err != nil {
			ch <- result{err: err}
			return
		}
		get := action.NewGet(cfg)
		get.Version = r.version
		rel, err := get.Run(r.name)
		ch <- result{rel: rel, err: err}
	}()

	select {
	case <-ctx.Done():
		return helmSummary{}, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return helmSummary{}, res.err
		}
		return summarize(res.rel), nil
	}
}

func summarize(rel *release.Release) helmSummary {
	var sum helmSummary
	if rel == nil {
		return sum
	}
	if rel.Chart != nil && rel.Chart.Metadata != nil {
		sum.chart = rel.Chart.Metadata.Name
		sum.version = rel.Chart.Metadata.Version
		sum.appVersion = rel.Chart.Metadata.AppVersion
	}
	if rel.Info != nil {
		sum.deployed = rel.Info.LastDeployed.Time
	}

	return sum
}
