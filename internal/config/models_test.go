package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestModelSettingsLoad(t *testing.T) {
	cfg := config.NewCustomModel()

	assert.Nil(t, cfg.Load("testdata/model_settings.yml"))
	assert.Equal(t, 1, len(cfg.K9s.Models))
	assert.Equal(t, 1, len(cfg.K9s.Models["v1/nodes"].Columns))

	assert.Equal(t, "metadata.labels['failure-domain.beta.kubernetes.io/zone']", cfg.K9s.Models["v1/nodes"].Columns[0].FieldPath)
}
