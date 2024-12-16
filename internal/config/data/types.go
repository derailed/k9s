// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"net/http"
	"net/url"
	"os"

	"github.com/derailed/k9s/internal/config/json"
	"k8s.io/client-go/tools/clientcmd/api"
)

// JSONValidator validate yaml configurations.
var JSONValidator = json.NewValidator()

const (
	// DefaultDirMod default unix perms for k9s directory.
	DefaultDirMod os.FileMode = 0744

	// DefaultFileMod default unix perms for k9s files.
	DefaultFileMod os.FileMode = 0600

	// MainConfigFile track main configuration file.
	MainConfigFile = "config.yaml"
)

// KubeSettings exposes kubeconfig context information.
type KubeSettings interface {
	// CurrentContextName returns the name of the current context.
	CurrentContextName() (string, error)

	// CurrentClusterName returns the name of the current cluster.
	CurrentClusterName() (string, error)

	// CurrentNamespaceName returns the name of the current namespace.
	CurrentNamespaceName() (string, error)

	// ContextNames returns all available context names.
	ContextNames() (map[string]struct{}, error)

	// CurrentContext returns the current context configuration.
	CurrentContext() (*api.Context, error)

	// GetContext returns a given context configuration or err if not found.
	GetContext(string) (*api.Context, error)

	// SetProxy sets the proxy for the active context, if present
	SetProxy(proxy func(*http.Request) (*url.URL, error))
}
