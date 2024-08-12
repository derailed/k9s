// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/config"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/cache"
)

const (
	k9sGitURL       = "https://api.github.com/repos/derailed/k9s/releases/latest"
	cacheSize       = 10
	cacheExpiry     = 1 * time.Hour
	k9sLatestRevKey = "k9sRev"
)

// ClusterInfoListener registers a listener for model changes.
type ClusterInfoListener interface {
	// ClusterInfoChanged notifies the cluster meta was changed.
	ClusterInfoChanged(prev, curr ClusterMeta)

	// ClusterInfoUpdated notifies the cluster meta was updated.
	ClusterInfoUpdated(ClusterMeta)
}

// ClusterMeta represents cluster meta data.
type ClusterMeta struct {
	Context, Cluster    string
	User                string
	K9sVer, K9sLatest   string
	K8sVer              string
	Cpu, Mem, Ephemeral int
}

// NewClusterMeta returns a new instance.
func NewClusterMeta() ClusterMeta {
	return ClusterMeta{
		Context:   client.NA,
		Cluster:   client.NA,
		User:      client.NA,
		K9sVer:    client.NA,
		K8sVer:    client.NA,
		Cpu:       0,
		Mem:       0,
		Ephemeral: 0,
	}
}

// Deltas diffs cluster meta return true if different, false otherwise.
func (c ClusterMeta) Deltas(n ClusterMeta) bool {
	if c.Cpu != n.Cpu || c.Mem != n.Mem || c.Ephemeral != n.Ephemeral {
		return true
	}

	return c.Context != n.Context ||
		c.Cluster != n.Cluster ||
		c.User != n.User ||
		c.K8sVer != n.K8sVer ||
		c.K9sVer != n.K9sVer ||
		c.K9sLatest != n.K9sLatest
}

// ClusterInfo models cluster metadata.
type ClusterInfo struct {
	cluster   *Cluster
	factory   dao.Factory
	data      ClusterMeta
	version   string
	cfg       *config.K9s
	listeners []ClusterInfoListener
	cache     *cache.LRUExpireCache
	mx        sync.RWMutex
}

// NewClusterInfo returns a new instance.
func NewClusterInfo(f dao.Factory, v string, cfg *config.K9s) *ClusterInfo {
	c := ClusterInfo{
		factory: f,
		cluster: NewCluster(f),
		data:    NewClusterMeta(),
		version: v,
		cfg:     cfg,
		cache:   cache.NewLRUExpireCache(cacheSize),
	}

	return &c
}

func (c *ClusterInfo) fetchK9sLatestRev() string {
	rev, ok := c.cache.Get(k9sLatestRevKey)
	if ok {
		return rev.(string)
	}

	latestRev, err := fetchLatestRev()
	if err != nil {
		log.Warn().Msgf("k9s latest rev fetch failed %s", err)
	} else {
		c.cache.Add(k9sLatestRevKey, latestRev, cacheExpiry)
	}

	return latestRev
}

// Reset resets context and reload.
func (c *ClusterInfo) Reset(f dao.Factory) {
	if f == nil {
		return
	}

	c.mx.Lock()
	{
		c.cluster, c.data = NewCluster(f), NewClusterMeta()
	}
	c.mx.Unlock()

	c.Refresh()
}

// Refresh fetches the latest cluster meta.
func (c *ClusterInfo) Refresh() {
	data := NewClusterMeta()
	if c.factory.Client().ConnectionOK() {
		data.Context = c.cluster.ContextName()
		data.Cluster = c.cluster.ClusterName()
		data.User = c.cluster.UserName()
		data.K8sVer = c.cluster.Version()
		ctx, cancel := context.WithTimeout(context.Background(), c.cluster.factory.Client().Config().CallTimeout())
		defer cancel()
		var mx client.ClusterMetrics
		if err := c.cluster.Metrics(ctx, &mx); err == nil {
			data.Cpu, data.Mem, data.Ephemeral = mx.PercCPU, mx.PercMEM, mx.PercEphemeral
		}
	}
	data.K9sVer = c.version
	v1 := NewSemVer(data.K9sVer)

	var latestRev string
	if !c.cfg.SkipLatestRevCheck {
		latestRev = c.fetchK9sLatestRev()
	}
	v2 := NewSemVer(latestRev)

	data.K9sVer, data.K9sLatest = v1.String(), v2.String()
	if v1.IsCurrent(v2) {
		data.K9sLatest = ""
	}

	if c.data.Deltas(data) {
		c.fireMetaChanged(c.data, data)
	} else {
		c.fireNoMetaChanged(data)
	}
	c.mx.Lock()
	{
		c.data = data
	}
	c.mx.Unlock()
}

// AddListener adds a new model listener.
func (c *ClusterInfo) AddListener(l ClusterInfoListener) {
	c.listeners = append(c.listeners, l)
}

// RemoveListener delete a listener from the list.
func (c *ClusterInfo) RemoveListener(l ClusterInfoListener) {
	victim := -1
	for i, lis := range c.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		c.listeners = append(c.listeners[:victim], c.listeners[victim+1:]...)
	}
}

func (c *ClusterInfo) fireMetaChanged(prev, cur ClusterMeta) {
	for _, l := range c.listeners {
		l.ClusterInfoChanged(prev, cur)
	}
}

func (c *ClusterInfo) fireNoMetaChanged(data ClusterMeta) {
	for _, l := range c.listeners {
		l.ClusterInfoUpdated(data)
	}
}

// Helpers...

func fetchLatestRev() (string, error) {
	log.Debug().Msgf("Fetching latest k9s rev...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k9sGitURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	m := make(map[string]interface{}, 20)
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}

	if v, ok := m["name"]; ok {
		log.Debug().Msgf("K9s latest rev: %q", v.(string))
		return v.(string), nil
	}

	return "", errors.New("No version found")
}
