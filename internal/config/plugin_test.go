// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
)

var pluginYmlTestData = Plugin{
	Scopes:      []string{"po", "dp"},
	Args:        []string{"-n", "$NAMESPACE", "-boolean"},
	ShortCut:    "shift-s",
	Description: "blee",
	Command:     "duh",
	Confirm:     true,
	Background:  false,
}

var test1YmlTestData = Plugin{
	Scopes:          []string{"po", "dp"},
	Args:            []string{"-n", "$NAMESPACE", "-boolean"},
	ShortCut:        "shift-s",
	Description:     "blee",
	Command:         "duh",
	Confirm:         true,
	Background:      false,
	OverwriteOutput: true,
}

var test2YmlTestData = Plugin{
	Scopes:          []string{"svc", "ing"},
	Args:            []string{"-n", "$NAMESPACE", "-oyaml"},
	ShortCut:        "shift-r",
	Description:     "bla",
	Command:         "duha",
	Confirm:         false,
	Background:      true,
	OverwriteOutput: false,
}

func TestPluginLoad(t *testing.T) {
	AppPluginsFile = "/tmp/k9s-test/fred.yaml"
	os.Setenv("XDG_DATA_HOME", "/tmp/k9s-test")
	xdg.Reload()

	p := NewPlugins()
	assert.NoError(t, p.Load("testdata/plugins.yaml"))

	assert.Equal(t, 1, len(p.Plugins))
	k, ok := p.Plugins["blah"]
	assert.True(t, ok)
	assert.ObjectsAreEqual(pluginYmlTestData, k)
}

func TestSinglePluginFileLoad(t *testing.T) {
	p := NewPlugins()
	assert.NoError(t, p.load("testdata/plugins.yaml"))
	assert.NoError(t, p.loadPluginDir("/random/dir/not/exist"))

	assert.Equal(t, 1, len(p.Plugins))
	k, ok := p.Plugins["blah"]
	assert.True(t, ok)

	assert.ObjectsAreEqual(pluginYmlTestData, k)
}

func TestMultiplePluginFilesLoad(t *testing.T) {
	p := NewPlugins()
	assert.NoError(t, p.load("testdata/plugins.yaml"))
	assert.NoError(t, p.loadPluginDir("testdata/plugins"))

	testPlugins := map[string]Plugin{
		"blah":  pluginYmlTestData,
		"test1": test1YmlTestData,
		"test2": test2YmlTestData,
	}

	assert.Equal(t, len(testPlugins), len(p.Plugins))
	for name, expectedPlugin := range testPlugins {
		k, ok := p.Plugins[name]
		assert.True(t, ok)
		assert.ObjectsAreEqual(expectedPlugin, k)
	}
}
