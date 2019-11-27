package watch

// BOZO!!
// import (
// 	v1beta1 "github.com/derailed/k9s/internal/informers/metrics/v1beta1"
// 	"github.com/derailed/k9s/internal/k9s"
// 	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
// )

// // Interface provides access to each of this group's versions.
// type Interface interface {
// 	// V1beta1 provides access to shared informers for resources in V1beta1.
// 	V1beta1() v1beta1.Interface
// }

// type SharedFactory interface {
// 	internalinterfaces.SharedInformerFactory
// 	Client() k9s.Connection
// }

// type group struct {
// 	factory          SharedFactory
// 	namespace        string
// 	tweakListOptions internalinterfaces.TweakListOptionsFunc
// }

// // New returns a new Interface.
// func New(f SharedFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
// 	return &group{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
// }

// // V1beta1 returns a new v1beta1.Interface.
// func (g *group) V1beta1() v1beta1.Interface {
// 	return v1beta1.New(g.factory, g.namespace, g.tweakListOptions)
// }
