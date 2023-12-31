// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"fmt"
	"io"
	"os"

	"github.com/derailed/k9s/internal/client"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Config tracks a context configuration.
type Config struct {
	Context *Context `yaml:"k9s"`
}

func NewConfig(ct *api.Context) *Config {
	return &Config{
		Context: NewContextFromConfig(ct),
	}
}

func (c *Config) Validate(conn client.Connection, ks KubeSettings) {
	c.Context.Validate(conn, ks)
}

func (c *Config) Dump(w io.Writer) {
	bb, _ := yaml.Marshal(&c)

	fmt.Fprintf(w, "%s\n", string(bb))
}

func (c *Config) Save(path string) error {
	if err := EnsureDirPath(path, DefaultDirMod); err != nil {
		return err
	}

	cfg, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, cfg, DefaultFileMod)
}
