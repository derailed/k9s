// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func newConfigForFile(path string) *client.Config {
	flags := genericclioptions.NewConfigFlags(client.UsePersistentConfig)
	flags.KubeConfig = &path

	return client.NewConfig(flags)
}

func TestCheckFatalConfigError(t *testing.T) {
	tests := map[string]struct {
		kubeconfig string
		wantFatal  bool
	}{
		"broken-duplicate-user": {
			kubeconfig: "testdata/bad-kubeconfig.yaml",
			wantFatal:  true,
		},
		"valid-config": {
			kubeconfig: "testdata/good-kubeconfig.yaml",
			wantFatal:  false,
		},
	}

	for k := range tests {
		u := tests[k]
		t.Run(k, func(t *testing.T) {
			cfg := newConfigForFile(u.kubeconfig)
			fatal := checkFatalConfigError(cfg)
			if u.wantFatal {
				assert.NotNil(t, fatal)
				assert.Equal(t, u.kubeconfig, fatal.path)
			} else {
				assert.Nil(t, fatal)
			}
		})
	}
}

func TestCheckFatalConfigErrorNil(t *testing.T) {
	assert.Nil(t, checkFatalConfigError(nil))
}

func TestFatalConfigErrorMessages(t *testing.T) {
	cause := errors.New(`error loading config file "/tmp/bad.yaml": duplicate name "default" in list`)
	e := &fatalConfigError{path: "/tmp/bad.yaml", err: cause}

	assert.ErrorIs(t, e, cause)
	assert.Contains(t, e.Error(), "/tmp/bad.yaml")
	assert.Contains(t, e.Error(), "could not be loaded")

	msg := e.UserMessage()
	assert.Contains(t, msg, "Unable to load your Kubernetes configuration")
	assert.Contains(t, msg, "File: /tmp/bad.yaml")
	assert.Contains(t, msg, `duplicate name "default"`)
	assert.Contains(t, msg, "K9s cannot start")
}

func TestFatalConfigErrorMessageNoPath(t *testing.T) {
	e := &fatalConfigError{err: errors.New("boom")}
	assert.Equal(t, "kubeconfig could not be loaded: boom", e.Error())
	assert.NotContains(t, e.UserMessage(), "File:")
}

func TestRootCause(t *testing.T) {
	tests := map[string]struct {
		err  error
		want string
	}{
		"nil": {
			err:  nil,
			want: "",
		},
		"joined-picks-config-load-line": {
			err: errors.Join(
				errors.New("some unrelated wrap"),
				errors.New(`error loading config file "/tmp/x.yaml": duplicate name "default" in list`),
			),
			want: `error loading config file "/tmp/x.yaml": duplicate name "default" in list`,
		},
		"plain": {
			err:  errors.New("cannot connect to context: foo"),
			want: "cannot connect to context: foo",
		},
	}

	for k := range tests {
		u := tests[k]
		t.Run(k, func(t *testing.T) {
			got := rootCause(u.err)
			assert.True(t, strings.Contains(got, u.want) || got == u.want)
		})
	}
}
