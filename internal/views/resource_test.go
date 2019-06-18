package views

// import (
// 	"context"
// 	"testing"

// 	"github.com/derailed/k9s/internal/config"
// 	"github.com/derailed/k9s/internal/resource"
// )

// func TestNewResource(t *testing.T) {
// 	mc := NewMockConnection()
// 	mk := NewMockKubeSettings()

// 	c := config.NewConfig(mk)
// 	c.SetConnection(mc)
// 	a := NewApp(c)
// 	l := resource.NewPodList(nil, "")
// 	v := newResourceView("fred", a, l)

// 	ctx, _ := context.WithCancel(context.Background())
// 	v.init(ctx, "")
// }
