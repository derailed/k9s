// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Dir struct {
	root string
	conn client.Connection
	ks   KubeSettings
}

func NewDir(root string, conn client.Connection, ks KubeSettings) *Dir {
	return &Dir{
		root: root,
		ks:   ks,
		conn: conn,
	}
}

func (d Dir) Load(n string, ct *api.Context) (*Config, error) {
	if ct == nil {
		return nil, errors.New("api.Context must not be nil")
	}

	var (
		path = filepath.Join(
			d.root,
			SanitizeFileName(ct.Cluster),
			SanitizeFileName(n),
			MainConfigFile,
		)
		cfg *Config
		err error
	)
	if f, e := os.Stat(path); os.IsNotExist(e) || f.Size() == 0 {
		log.Debug().Msgf("Context config not found! Generating... %q", path)
		cfg, err = d.genConfig(path, ct)
	} else {
		log.Debug().Msgf("Found existing context config: %q", path)
		cfg, err = d.loadConfig(path)
	}

	return cfg, err
}

func (d *Dir) genConfig(path string, ct *api.Context) (*Config, error) {
	cfg := NewConfig(ct)
	cfg.Validate(d.conn, d.ks)
	if err := cfg.Save(path); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (d *Dir) loadConfig(path string) (*Config, error) {
	bb, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(bb, &cfg); err != nil {
		return nil, err
	}
	cfg.Validate(d.conn, d.ks)

	return &cfg, nil
}
