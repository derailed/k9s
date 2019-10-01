package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAliasesLoad(t *testing.T) {
	aa := config.NewAliases()
	assert.Nil(t, aa.LoadAliases("test_assets/alias.yml"))

	assert.Equal(t, 27, len(aa.Alias))
}

func TestAliasesSave(t *testing.T) {
	aa := config.NewAliases()

	aa.Alias["test"] = "fred"
	aa.Alias["blee"] = "duh"
	aa.SaveAliases("/tmp/a.yml")

	assert.Nil(t, aa.LoadAliases("/tmp/a.yml"))
	assert.Equal(t, 28, len(aa.Alias))
}
