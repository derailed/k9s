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

	uu := []struct {
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

	for i := range uu {
		u := uu[i]
		t.Run(u.name, func(t *testing.T) {
			configAlias := config.NewAliases()
			for _, aliases := range u.aliases {
				for _, a := range aliases.aliases {
					configAlias.Define(aliases.cmd, a)
				}
			}
			for alias, cmd := range u.registeredCommands {
				v, ok := configAlias.Get(alias)
				assert.True(t, ok)
				assert.Equal(t, cmd, v, "Wrong command for alias "+alias)
			}
		})
	}
}

func TestAliasesLoad(t *testing.T) {
	a := config.NewAliases()

	assert.Nil(t, a.LoadAliases("testdata/alias.yml"))
	assert.Equal(t, 2, len(a.Alias))
}

func TestAliasesSave(t *testing.T) {
	a := config.NewAliases()
	a.Alias["test"] = "fred"
	a.Alias["blee"] = "duh"

	assert.Nil(t, a.SaveAliases("/tmp/a.yml"))
	assert.Nil(t, a.LoadAliases("/tmp/a.yml"))
	assert.Equal(t, 2, len(a.Alias))
}
