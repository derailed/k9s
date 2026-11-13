// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func helmSecret(ns, name, version, status, uid, rv string) metav1.PartialObjectMetadata {
	return metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ns,
			UID:             types.UID(uid),
			ResourceVersion: rv,
			Labels: map[string]string{
				"owner":   "helm",
				"name":    name,
				"version": version,
				"status":  status,
			},
		},
	}
}

func TestLatestHelmRevs(t *testing.T) {
	items := []metav1.PartialObjectMetadata{
		helmSecret("ns1", "app", "1", "superseded", "u1", "10"),
		helmSecret("ns1", "app", "3", "deployed", "u3", "30"),
		helmSecret("ns1", "app", "2", "superseded", "u2", "20"),
		helmSecret("ns2", "app", "1", "failed", "u4", "40"),
		helmSecret("ns1", "other", "12", "deployed", "u5", "50"),
		helmSecret("ns1", "other", "9", "superseded", "u6", "60"),
		helmSecret("ns1", "bad", "nope", "deployed", "u7", "70"),
		{ObjectMeta: metav1.ObjectMeta{Namespace: "ns1", UID: "u8"}},
	}

	rr := latestHelmRevs(items)
	sort.Slice(rr, func(i, j int) bool {
		return rr[i].ns+"/"+rr[i].name < rr[j].ns+"/"+rr[j].name
	})

	assert.Len(t, rr, 3)
	assert.Equal(t, "ns1/app", rr[0].ns+"/"+rr[0].name)
	assert.Equal(t, 3, rr[0].version)
	assert.Equal(t, "deployed", rr[0].status)
	assert.Equal(t, "u3/30", rr[0].key())
	assert.Equal(t, "ns1/other", rr[1].ns+"/"+rr[1].name)
	assert.Equal(t, 12, rr[1].version)
	assert.Equal(t, "ns2/app", rr[2].ns+"/"+rr[2].name)
	assert.Equal(t, 1, rr[2].version)
}

func TestHelmRevToRelease(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	r := helmRev{ns: "ns1", name: "app", status: "deployed", version: 4, created: created}

	rel := r.toRelease(nil)
	assert.Equal(t, "app", rel.Name)
	assert.Equal(t, "ns1", rel.Namespace)
	assert.Equal(t, 4, rel.Version)
	assert.Equal(t, "deployed", rel.Info.Status.String())
	assert.Equal(t, created, rel.Info.LastDeployed.Time)
	assert.Nil(t, rel.Chart)

	deployed := created.Add(time.Hour)
	rel = r.toRelease(&helmSummary{chart: "nginx", version: "1.2.3", appVersion: "9", deployed: deployed})
	assert.Equal(t, "nginx", rel.Chart.Metadata.Name)
	assert.Equal(t, "1.2.3", rel.Chart.Metadata.Version)
	assert.Equal(t, "9", rel.Chart.Metadata.AppVersion)
	assert.Equal(t, deployed, rel.Info.LastDeployed.Time)
}

func TestHelmSummariesClaimAndFinish(t *testing.T) {
	c := newHelmSummaries()

	assert.True(t, c.claim("a"))
	assert.False(t, c.claim("a"), "already in flight")

	c.finish("a", "ns1", helmSummary{chart: "x"}, nil)
	sum, ok := c.get("a")
	assert.True(t, ok)
	assert.Equal(t, "x", sum.chart)

	assert.True(t, c.claim("b"))
	c.finish("b", "ns1", helmSummary{}, assert.AnError)
	_, ok = c.get("b")
	assert.False(t, ok)
	assert.False(t, c.claim("b"), "backing off after failure")

	c.mx.Lock()
	c.failed["b"] = time.Now().Add(-2 * helmFailureBackoff)
	c.mx.Unlock()
	assert.True(t, c.claim("b"), "backoff expired")
}

func TestHelmSummariesPrune(t *testing.T) {
	c := newHelmSummaries()
	c.done["keep1"] = helmCacheEntry{ns: "ns1"}
	c.done["drop1"] = helmCacheEntry{ns: "ns1"}
	c.done["other"] = helmCacheEntry{ns: "ns2"}

	c.prune("ns1", map[string]struct{}{"keep1": {}})
	_, ok := c.get("keep1")
	assert.True(t, ok)
	_, ok = c.get("drop1")
	assert.False(t, ok)
	_, ok = c.get("other")
	assert.True(t, ok, "other namespaces are out of scope")

	c.prune("", map[string]struct{}{"keep1": {}})
	_, ok = c.get("other")
	assert.False(t, ok)
}

func TestSummarize(t *testing.T) {
	assert.Equal(t, helmSummary{}, summarize(nil))
}
