// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/derailed/k9s/internal/slogs"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MultiConnection wraps multiple connections to allow viewing resources across contexts.
type MultiConnection struct {
	configs     map[string]*Config
	connections map[string]Connection
	contexts    []string // ordered list of context names
	primary     string   // primary context (first selected)
	mu          sync.RWMutex
	log         *slog.Logger
}

// NewMultiConnection creates a new multi-context connection.
func NewMultiConnection(contextNames []string, baseConfig *Config, log *slog.Logger) (*MultiConnection, error) {
	if len(contextNames) == 0 {
		return nil, errors.New("at least one context must be specified")
	}

	mc := &MultiConnection{
		configs:     make(map[string]*Config),
		connections: make(map[string]Connection),
		contexts:    contextNames,
		primary:     contextNames[0],
		log:         log.With(slogs.Subsys, "multi-connection"),
	}

	// Initialize connections for each context
	rawConfig, err := baseConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	for _, ctx := range contextNames {
		if _, ok := rawConfig.Contexts[ctx]; !ok {
			return nil, fmt.Errorf("context %q not found in kubeconfig", ctx)
		}

		// Clone the config for this context
		cfg := NewConfig(baseConfig.Flags())
		if err := cfg.SwitchContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to switch to context %q: %w", ctx, err)
		}

		// Initialize connection for this context
		conn, err := InitConnection(cfg, log)
		if err != nil {
			mc.log.Warn("Failed to initialize connection", slogs.Context, ctx, slogs.Error, err)
			// Continue with other contexts even if one fails
		}

		mc.configs[ctx] = cfg
		mc.connections[ctx] = conn
	}

	if len(mc.connections) == 0 {
		return nil, errors.New("failed to initialize any connections")
	}

	return mc, nil
}

// Contexts returns the list of active context names.
func (mc *MultiConnection) Contexts() []string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return append([]string{}, mc.contexts...)
}

// IsMultiContext returns true if managing multiple contexts.
func (mc *MultiConnection) IsMultiContext() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return len(mc.contexts) > 1
}

// PrimaryContext returns the primary (first) context name.
func (mc *MultiConnection) PrimaryContext() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.primary
}

// Connection interface implementation

// Config returns the primary config.
func (mc *MultiConnection) Config() *Config {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.configs[mc.primary]
}

// ConnectionOK checks if at least one connection is OK.
func (mc *MultiConnection) ConnectionOK() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, conn := range mc.connections {
		if conn.ConnectionOK() {
			return true
		}
	}
	return false
}

// Dial connects to the primary context's API server.
func (mc *MultiConnection) Dial() (kubernetes.Interface, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return nil, fmt.Errorf("primary context %q not found", mc.primary)
	}
	return conn.Dial()
}

// DialLogs connects to the primary context's API server for logs.
func (mc *MultiConnection) DialLogs() (kubernetes.Interface, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return nil, fmt.Errorf("primary context %q not found", mc.primary)
	}
	return conn.DialLogs()
}

// SwitchContext switches all connections to a single context.
// This effectively exits multi-context mode.
func (mc *MultiConnection) SwitchContext(ctx string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Clear all contexts and set to single context
	mc.contexts = []string{ctx}
	mc.primary = ctx

	// Check if we already have this connection
	if conn, ok := mc.connections[ctx]; ok {
		// Keep only this connection
		mc.connections = map[string]Connection{ctx: conn}
		mc.configs = map[string]*Config{ctx: mc.configs[ctx]}
		return nil
	}

	// Create new connection for this context
	cfg := NewConfig(mc.Config().Flags())
	if err := cfg.SwitchContext(ctx); err != nil {
		return err
	}

	conn, err := InitConnection(cfg, mc.log)
	if err != nil {
		return err
	}

	mc.connections = map[string]Connection{ctx: conn}
	mc.configs = map[string]*Config{ctx: cfg}

	return nil
}

// CachedDiscovery connects to the primary context's cached discovery client.
func (mc *MultiConnection) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return nil, fmt.Errorf("primary context %q not found", mc.primary)
	}
	return conn.CachedDiscovery()
}

// RestConfig connects to the primary context's REST client.
func (mc *MultiConnection) RestConfig() (*restclient.Config, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return nil, fmt.Errorf("primary context %q not found", mc.primary)
	}
	return conn.RestConfig()
}

// MXDial connects to the primary context's metrics server.
func (mc *MultiConnection) MXDial() (*versioned.Clientset, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return nil, fmt.Errorf("primary context %q not found", mc.primary)
	}
	return conn.MXDial()
}

// DynDial connects to the primary context's dynamic client.
func (mc *MultiConnection) DynDial() (dynamic.Interface, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return nil, fmt.Errorf("primary context %q not found", mc.primary)
	}
	return conn.DynDial()
}

// HasMetrics checks if the primary context has metrics server.
func (mc *MultiConnection) HasMetrics() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return false
	}
	return conn.HasMetrics()
}

// ValidNamespaceNames returns namespaces from all contexts.
func (mc *MultiConnection) ValidNamespaceNames() (NamespaceNames, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	allNamespaces := make(NamespaceNames)
	for _, conn := range mc.connections {
		nss, err := conn.ValidNamespaceNames()
		if err != nil {
			mc.log.Warn("Failed to get namespaces", slogs.Error, err)
			continue
		}
		for ns := range nss {
			allNamespaces[ns] = struct{}{}
		}
	}

	return allNamespaces, nil
}

// IsValidNamespace checks if the namespace exists in any context.
func (mc *MultiConnection) IsValidNamespace(ns string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, conn := range mc.connections {
		if conn.IsValidNamespace(ns) {
			return true
		}
	}
	return false
}

// ServerVersion returns the primary context's server version.
func (mc *MultiConnection) ServerVersion() (*version.Info, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return nil, fmt.Errorf("primary context %q not found", mc.primary)
	}
	return conn.ServerVersion()
}

// CheckConnectivity checks if at least one context is reachable.
func (mc *MultiConnection) CheckConnectivity() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, conn := range mc.connections {
		if conn.CheckConnectivity() {
			return true
		}
	}
	return false
}

// ActiveContext returns a string representing active contexts.
func (mc *MultiConnection) ActiveContext() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(mc.contexts) == 1 {
		return mc.contexts[0]
	}
	return fmt.Sprintf("multi:%d", len(mc.contexts))
}

// ActiveNamespace returns the primary context's active namespace.
func (mc *MultiConnection) ActiveNamespace() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return BlankNamespace
	}
	return conn.ActiveNamespace()
}

// IsActiveNamespace checks against the primary context.
func (mc *MultiConnection) IsActiveNamespace(ns string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	conn, ok := mc.connections[mc.primary]
	if !ok {
		return false
	}
	return conn.IsActiveNamespace(ns)
}

// CanI checks authorization across all contexts.
// Returns true only if authorized in all contexts.
func (mc *MultiConnection) CanI(ns string, gvr *GVR, n string, verbs []string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Must have access in all contexts
	for ctx, conn := range mc.connections {
		can, err := conn.CanI(ns, gvr, n, verbs)
		if err != nil {
			mc.log.Debug("Authorization check failed", slogs.Context, ctx, slogs.Error, err)
			return false, err
		}
		if !can {
			return false, nil
		}
	}

	return true, nil
}

// ForEachConnection executes a function for each connection.
// Useful for operations that need to aggregate data from all contexts.
func (mc *MultiConnection) ForEachConnection(fn func(ctx string, conn Connection) error) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, ctx := range mc.contexts {
		conn, ok := mc.connections[ctx]
		if !ok {
			continue
		}
		if err := fn(ctx, conn); err != nil {
			return fmt.Errorf("context %q: %w", ctx, err)
		}
	}

	return nil
}
