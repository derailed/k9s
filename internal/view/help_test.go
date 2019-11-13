package view_test

// import (
// 	"testing"

// 	"github.com/derailed/k9s/internal/config"
// 	"github.com/derailed/k9s/internal/model"
// 	"github.com/stretchr/testify/assert"
// 	v1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func newNS(n string) v1.Namespace {
// 	return v1.Namespace{ObjectMeta: metav1.ObjectMeta{
// 		Name: n,
// 	}}
// }

// func TestHelpNew(t *testing.T) {
// 	a := view.NewApp(config.NewConfig(ks{}))
// 	v := view.NewHelp()
// 	ctx := context.WithValue(ui.KeyApp, app)
// 	v.Init(ctx)

// 	app.SetHints(model.MenuHints{{Mnemonic: "blee", Description: "duh"}})

// 	assert.Equal(t, "<blee>", v.GetCell(1, 0).Text)
// 	assert.Equal(t, "duh", v.GetCell(1, 1).Text)
// }
