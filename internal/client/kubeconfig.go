// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"reflect"
	"slices"

	"github.com/derailed/k9s/internal/slogs"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// K9s picks up kubeconfig edits made while it runs, but hand picks what it
// applies so that the shell and K9s keep their own notion of the current
// context:
//   - Context listings always reflect the latest kubeconfig content.
//   - The active context stays pinned. A current-context change made outside
//     of K9s, e.g. via kubectl config use-context, never switches K9s.
//   - Changed credentials or server settings for the active context are
//     applied as long as the context still exists and points at the same
//     cluster. Otherwise K9s keeps the settings it loaded and reports the
//     active context as detached until the user switches context.

// fileStamp captures the state of a kubeconfig file to detect changes.
type fileStamp struct {
	path    string
	exists  bool
	size    int64
	modTime int64
}

func stampFiles(paths []string) []fileStamp {
	ss := make([]fileStamp, 0, len(paths))
	for _, p := range paths {
		s := fileStamp{path: p}
		info, err := os.Stat(p)
		switch {
		case err == nil:
			s.exists, s.size, s.modTime = true, info.Size(), info.ModTime().UnixNano()
		case !errors.Is(err, fs.ErrNotExist):
			slog.Debug("Kubeconfig stat failed", slogs.Path, p, slogs.Error, err)
		}
		ss = append(ss, s)
	}

	return ss
}

// kubeconfigFiles returns the files the kubeconfig gets merged from.
func kubeconfigFiles(f *genericclioptions.ConfigFlags) []string {
	if isSet(f.KubeConfig) {
		return []string{*f.KubeConfig}
	}

	return clientcmd.NewDefaultClientConfigLoadingRules().GetLoadingPrecedence()
}

// cloneFlags copies all settings so the clone reloads the kubeconfig from disk.
func cloneFlags(f *genericclioptions.ConfigFlags) *genericclioptions.ConfigFlags {
	c := genericclioptions.NewConfigFlags(UsePersistentConfig)
	c.CacheDir, c.KubeConfig = f.CacheDir, f.KubeConfig
	c.ClusterName, c.AuthInfoName, c.Context = f.ClusterName, f.AuthInfoName, f.Context
	c.Namespace, c.APIServer, c.TLSServerName = f.Namespace, f.APIServer, f.TLSServerName
	c.Insecure, c.CertFile, c.KeyFile, c.CAFile = f.Insecure, f.CertFile, f.KeyFile, f.CAFile
	c.BearerToken = f.BearerToken
	c.Impersonate, c.ImpersonateUID = f.Impersonate, f.ImpersonateUID
	c.ImpersonateGroup, c.ImpersonateUserExtra = f.ImpersonateGroup, f.ImpersonateUserExtra
	c.Username, c.Password = f.Username, f.Password
	c.Timeout, c.DisableCompression, c.ProxyURL = f.Timeout, f.DisableCompression, f.ProxyURL
	c.WrapConfigFn = f.WrapConfigFn

	return c
}

// refresh reloads the kubeconfig when any of its files changed on disk.
func (c *Config) refresh() {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.flags == nil {
		return
	}
	stamps := stampFiles(kubeconfigFiles(c.flags))
	if !c.primed {
		c.prime(stamps)
		return
	}
	if slices.Equal(stamps, c.stamps) {
		return
	}
	c.stamps = stamps
	c.reload()
}

// prime records the kubeconfig K9s starts out with. Files are stamped before
// loading so an edit landing in between gets picked up by the next refresh.
func (c *Config) prime(stamps []fileStamp) {
	cfg, err := c.flags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		// Leave unprimed so the next refresh tries again.
		return
	}
	c.stamps, c.active, c.primed = stamps, cfg, true
}

func (c *Config) reload() {
	name := c.active.CurrentContext
	if isSet(c.flags.Context) {
		name = *c.flags.Context
	}
	next := cloneFlags(c.flags)
	if name != "" {
		next.Context = &name
	}
	cfg, err := next.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		slog.Warn("Kubeconfig reload failed. Keeping previous configuration", slogs.Error, err)
		return
	}
	slog.Debug("Kubeconfig reloaded", slogs.Context, name)
	c.latest = next
	if name == "" {
		// K9s has no active context, so there is nothing to carry over.
		return
	}

	if reason := detachReason(name, &c.active, &cfg); reason != "" {
		if reason != c.detached {
			slog.Warn("Kubeconfig no longer matches the active context. Keeping loaded settings",
				slogs.Context, name,
				slogs.Reason, reason,
			)
		}
		c.detached = reason
		return
	}
	if !sameConnection(name, c.flags, &c.active, &cfg) {
		c.revision++
		slog.Info("Kubeconfig changes applied to active context", slogs.Context, name)
	}
	c.flags, c.active, c.detached = next, cfg, ""
}

// detachReason explains why a reloaded kubeconfig can no longer back the
// active context or returns blank if it still can.
func detachReason(name string, prev, next *api.Config) string {
	ct, ok := next.Contexts[name]
	if !ok {
		return fmt.Sprintf("Active context %q was removed from the kubeconfig", name)
	}
	if was := prev.Contexts[name]; was != nil && was.Cluster != ct.Cluster {
		return fmt.Sprintf("Active context %q now points at cluster %q instead of %q", name, ct.Cluster, was.Cluster)
	}

	return ""
}

// sameConnection checks whether the active context still resolves to the same
// user and cluster settings.
func sameConnection(name string, f *genericclioptions.ConfigFlags, prev, next *api.Config) bool {
	cluster, user := connectionNames(name, f, prev)
	nextCluster, nextUser := connectionNames(name, f, next)
	if cluster != nextCluster || user != nextUser {
		return false
	}

	return reflect.DeepEqual(prev.Clusters[cluster], next.Clusters[cluster]) &&
		reflect.DeepEqual(prev.AuthInfos[user], next.AuthInfos[user])
}

// connectionNames returns the cluster and user a context resolves to while
// honoring CLI overrides.
func connectionNames(name string, f *genericclioptions.ConfigFlags, cfg *api.Config) (cluster, user string) {
	if ct, ok := cfg.Contexts[name]; ok {
		cluster, user = ct.Cluster, ct.AuthInfo
	}
	if isSet(f.ClusterName) {
		cluster = *f.ClusterName
	}
	if isSet(f.AuthInfoName) {
		user = *f.AuthInfoName
	}

	return cluster, user
}

// ActiveRevision returns a counter that increments every time kubeconfig
// changes to the active context's user or cluster settings get applied.
// Connections built before the increment use stale settings.
func (c *Config) ActiveRevision() uint64 {
	c.refresh()
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.revision
}

// DetachedReason explains why the kubeconfig on disk no longer backs the
// active context or returns blank when it does.
func (c *Config) DetachedReason() string {
	c.refresh()
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.detached
}
