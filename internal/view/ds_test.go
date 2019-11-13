package view_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/derailed/k9s/internal/config"
// 	"github.com/derailed/k9s/internal/resource"
// 	"github.com/derailed/k9s/internal/ui"
// 	"github.com/derailed/k9s/internal/view"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDaemonSet(t *testing.T) {
// 	l := resource.NewDaemonSetList(nil, "fred")
// 	v := view.NewDaemonSet("blee", "", l)
// 	ctx := context.WithValue(ui.KeyApp, NewApp(config.NewConfig(ks{})))
// 	v.Init(ctx)

// 	assert.Equal(t, 10, len(v.Hints()))
// }
