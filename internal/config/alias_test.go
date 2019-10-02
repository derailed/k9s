package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAliasDefine(t *testing.T) {
	uu := map[string]struct {
		aa []string
	}{
		"one":   {[]string{"blee", "duh"}},
		"multi": {[]string{"blee", "duh", "fred", "zorg"}},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			a := config.NewAliases()
			a.Define(u.aa...)
			for i := 0; i < len(u.aa); i += 2 {
				v, ok := a.Get(u.aa[i])

				assert.True(t, ok)
				assert.Equal(t, u.aa[i+1], v)
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
