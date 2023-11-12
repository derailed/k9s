package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

var pluginYmlTestData = config.Plugin{
	Scopes:      []string{"po", "dp"},
	Args:        []string{"-n", "$NAMESPACE", "-boolean"},
	ShortCut:    "shift-s",
	Description: "blee",
	Command:     "duh",
	Confirm:     true,
	Background:  false,
}

var test1YmlTestData = config.Plugin{
	Scopes:      []string{"po", "dp"},
	Args:        []string{"-n", "$NAMESPACE", "-boolean"},
	ShortCut:    "shift-s",
	Description: "blee",
	Command:     "duh",
	Confirm:     true,
	Background:  false,
}

var test2YmlTestData = config.Plugin{
	Scopes:      []string{"svc", "ing"},
	Args:        []string{"-n", "$NAMESPACE", "-oyaml"},
	ShortCut:    "shift-r",
	Description: "bla",
	Command:     "duha",
	Confirm:     false,
	Background:  true,
}

func TestSinglePluginFileLoad(t *testing.T) {
	p := config.NewPlugins()
	assert.Nil(t, p.LoadPlugins("testdata/plugin.yml", []string{"/random/dir/not/exist"}))

	assert.Equal(t, 1, len(p.Plugin))
	k, ok := p.Plugin["blah"]
	assert.True(t, ok)

	assert.ObjectsAreEqual(pluginYmlTestData, k)
}

func TestMultiplePluginFilesLoad(t *testing.T) {
	p := config.NewPlugins()
	assert.Nil(t, p.LoadPlugins("testdata/plugin.yml", []string{"testdata/plugins"}))

	testPlugins := map[string]config.Plugin{
		"blah":  pluginYmlTestData,
		"test1": test1YmlTestData,
		"test2": test2YmlTestData,
	}

	assert.Equal(t, len(testPlugins), len(p.Plugin))
	for name, expectedPlugin := range testPlugins {
		k, ok := p.Plugin[name]
		assert.True(t, ok)
		assert.ObjectsAreEqual(expectedPlugin, k)
	}
}
