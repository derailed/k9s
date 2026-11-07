// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"log/slog"
	"sync"

	"github.com/derailed/k9s/internal/slogs"
)

// kubeconfigTracker remembers which kubeconfig reload the app last applied.
type kubeconfigTracker struct {
	mx       sync.Mutex
	revision uint64
	detached string
}

// observe records the latest reload state and reports what changed.
func (k *kubeconfigTracker) observe(rev uint64, detached string) (revChanged, detachedChanged bool) {
	k.mx.Lock()
	defer k.mx.Unlock()

	revChanged, detachedChanged = rev != k.revision, detached != k.detached
	k.revision, k.detached = rev, detached

	return revChanged, detachedChanged
}

// reset starts tracking a freshly switched context.
func (k *kubeconfigTracker) reset(rev uint64) {
	k.mx.Lock()
	defer k.mx.Unlock()

	k.revision, k.detached = rev, ""
}

// applyKubeconfigChanges keeps the app in sync with kubeconfig edits made to
// the active context while K9s runs. It reports whether informers got
// restarted, in which case views must resubscribe, and whether the user got
// flashed.
func (a *App) applyKubeconfigChanges() (restarted, flashed bool) {
	cfg := a.Conn().Config()
	rev, detached := cfg.ActiveRevision(), cfg.DetachedReason()
	revChanged, detachedChanged := a.kubeconfig.observe(rev, detached)
	if detachedChanged && detached != "" {
		a.Flash().Warnf("%s. Keeping loaded settings until you switch context", detached)
		flashed = true
	}
	if !revChanged {
		return false, flashed
	}

	ct := a.Config.ActiveContextName()
	slog.Info("Applying kubeconfig changes", slogs.Context, ct)
	a.factory.RestartInformers(a.Config.ActiveNamespace())
	a.Flash().Infof("Kubeconfig changes applied to context %q", ct)

	return true, true
}
