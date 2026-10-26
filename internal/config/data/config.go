// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"fmt"
	"io"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Config tracks a context configuration.
type Config struct {
	Context *Context `yaml:"k9s"`
	mx      sync.RWMutex
}

// NewConfig returns a new config.
func NewConfig(ct *api.Context) *Config {
	return &Config{
		Context: NewContextFromConfig(ct),
	}
}

// Merge merges configs and updates receiver.
func (c *Config) Merge(c1 *Config) {
	if c1 == nil {
		return
	}
	if c.Context != nil && c1.Context != nil {
		c.Context.merge(c1.Context)
	}
}

// Validate ensures config is in norms.
func (c *Config) Validate(conn client.Connection, contextName, clusterName string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.Context == nil {
		c.Context = NewContext()
	}
	c.Context.Validate(conn, contextName, clusterName)
}

// Dump used for debugging.
func (c *Config) Dump(w io.Writer) {
	bb, _ := yaml.Marshal(&c)

	_, _ = fmt.Fprintf(w, "%s\n", string(bb))
}
