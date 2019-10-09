package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAliasDefine(t *testing.T) {
	type aliasDef struct {
		cmd     string
		aliases []string
	}

	tts := []struct {
		name               string
		aliases            []aliasDef
		registeredCommands map[string]string
	}{
		{
			name: "simple aliases",
			aliases: []aliasDef{
				{
					cmd:     "one",
					aliases: []string{"blee", "duh"},
				},
			},
			registeredCommands: map[string]string{
				"blee": "one",
				"duh":  "one",
			},
		},
		{
			name: "duplicated aliases",
			aliases: []aliasDef{
				{
					cmd:     "one",
					aliases: []string{"blee", "duh"},
				}, {
					cmd:     "two",
					aliases: []string{"blee", "duh", "fred", "zorg"},
				},
			},
			registeredCommands: map[string]string{
				"blee": "one",
				"duh":  "one",
				"fred": "two",
				"zorg": "two",
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			configAlias := config.NewAliases()
			for _, aliases := range tt.aliases {
				for _, a := range aliases.aliases {
					configAlias.Define(aliases.cmd, a)
				}
			}
			for alias, cmd := range tt.registeredCommands {
				v, ok := configAlias.Get(alias)
				assert.True(t, ok)
				assert.Equal(t, cmd, v, "Wrong command for alias "+alias)
			}
		})
	}
}

func TestAliasesLoad(t *testing.T) {
	a := config.NewAliases()
	assert.Nil(t, a.LoadAliases("test_assets/alias.yml"))

	assert.Equal(t, 27, len(a.Alias))
}

func TestAliasesSave(t *testing.T) {
	a := config.NewAliases()

	a.Alias["test"] = "fred"
	a.Alias["blee"] = "duh"
	a.SaveAliases("/tmp/a.yml")

	assert.Nil(t, a.LoadAliases("/tmp/a.yml"))
	assert.Equal(t, 28, len(a.Alias))
}
