package render_test

// BOZO!! revamp with latest...

// import (
// 	"testing"

// 	"github.com/derailed/k9s/internal/render"
// 	ofaas "github.com/openfaas/faas-provider/types"
// 	"github.com/stretchr/testify/assert"
// )

// func TestOpenFaasRender(t *testing.T) {
// 	c := render.OpenFaas{}
// 	r := render.NewRow(9)
// 	c.Render(makeFn("blee"), "", &r)

// 	assert.Equal(t, "default/blee", r.ID)
// 	assert.Equal(t, render.Fields{"default", "blee", "Ready", "nginx:0", "fred=blee", "10", "1", "1"}, r.Fields[:8])
// }

// // Helpers...

// func makeFn(n string) render.OpenFaasRes {
// 	return render.OpenFaasRes{
// 		Function: ofaas.FunctionStatus{
// 			Name:              n,
// 			Namespace:         "default",
// 			Image:             "nginx:0",
// 			InvocationCount:   10,
// 			Replicas:          1,
// 			AvailableReplicas: 1,
// 			Labels:            &map[string]string{"fred": "blee"},
// 		},
// 	}
// }
