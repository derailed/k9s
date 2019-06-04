package k8s

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	authorizationv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NA Not available
const NA = "n/a"

var supportedMetricsAPIVersions = []string{"v1beta1"}

type (
	// APIGroup represents a K8s resource descriptor.
	APIGroup struct {
		Resource             string
		Group, Kind, Version string
		Plural, Singular     string
		Aliases              []string
	}

	// Collection of empty interfaces.
	Collection []interface{}

	// Cruder represent a crudable Kubernetes resource.
	Cruder interface {
		Get(ns string, name string) (interface{}, error)
		List(ns string) (Collection, error)
		Delete(ns string, name string) error
		SetFieldSelector(string)
		SetLabelSelector(string)
	}

	// Connection represents a Kubenetes apiserver connection.
	Connection interface {
		Config() *Config
		DialOrDie() kubernetes.Interface
		SwitchContextOrDie(ctx string)
		NSDialOrDie() dynamic.NamespaceableResourceInterface
		RestConfigOrDie() *restclient.Config
		MXDial() (*versioned.Clientset, error)
		DynDialOrDie() dynamic.Interface
		HasMetrics() bool
		IsNamespaced(n string) bool
		SupportsResource(group string) bool
		ValidNamespaces() ([]v1.Namespace, error)
		NodePods(node string) (*v1.PodList, error)
		SupportsRes(grp string, versions []string) (string, bool, error)
		ServerVersion() (*version.Info, error)
		FetchNodes() (*v1.NodeList, error)
		CurrentNamespaceName() (string, error)
		CanIAccess(ns, name, resURL string, verbs []string) (bool, error)
	}

	// APIClient represents a Kubernetes api client.
	APIClient struct {
		config          *Config
		client          kubernetes.Interface
		dClient         dynamic.Interface
		nsClient        dynamic.NamespaceableResourceInterface
		mxsClient       *versioned.Clientset
		useMetricServer bool
		log             zerolog.Logger
		mx              sync.Mutex
	}
)

// InitConnectionOrDie initialize connection from command line args.
// Checks for connectivity with the api server.
func InitConnectionOrDie(config *Config, logger zerolog.Logger) *APIClient {
	conn := APIClient{config: config, log: logger}
	conn.useMetricServer = conn.supportsMxServer()

	return &conn
}

// CanIAccess checks if user has access to a certain resource.
func (a *APIClient) CanIAccess(ns, name, resURL string, verbs []string) (bool, error) {
	_, gr := schema.ParseResourceArg(strings.ToLower(resURL))
	sar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace:   ns,
				Group:       gr.Group,
				Resource:    gr.Resource,
				Subresource: "",
				Name:        name,
			},
		},
	}

	user, _ := a.Config().CurrentUserName()
	groups, _ := a.Config().CurrentGroupNames()
	log.Debug().Msgf("AuthInfo user/groups: %s:%v", user, groups)

	var (
		resp  *authorizationv1.SelfSubjectAccessReview
		err   error
		allow bool
	)
	for _, v := range verbs {
		sar.Spec.ResourceAttributes.Verb = v
		resp, err = a.DialOrDie().AuthorizationV1().SelfSubjectAccessReviews().Create(sar)
		if err != nil {
			log.Warn().Err(err).Msgf("CanIAccess")
			return false, err
		}
		log.Debug().Msgf("CHECKING ACCESS res:%s-%q for NS: %q Verb: %s -> %t, %s", resURL, name, ns, v, resp.Status.Allowed, resp.Status.Reason)
		if !resp.Status.Allowed {
			return false, err
		}
		allow = true
	}
	log.Debug().Msgf("GRANT ACCESS? >>> %t", allow)

	return allow, nil
}

// CurrentNamespaceName return namespace name set via either cli arg or cluster config.
func (a *APIClient) CurrentNamespaceName() (string, error) {
	return a.config.CurrentNamespaceName()
}

// ServerVersion returns the current server version info.
func (a *APIClient) ServerVersion() (*version.Info, error) {
	return a.DialOrDie().Discovery().ServerVersion()
}

// FetchNodes returns all available nodes.
func (a *APIClient) FetchNodes() (*v1.NodeList, error) {
	return a.DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
}

// ValidNamespaces returns all available namespaces.
func (a *APIClient) ValidNamespaces() ([]v1.Namespace, error) {
	nn, err := a.DialOrDie().CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nn.Items, nil
}

// NodePods returns a collection of all available pods on a given node.
func (a *APIClient) NodePods(node string) (*v1.PodList, error) {
	const selFmt = "spec.nodeName=%s,status.phase!=%s,status.phase!=%s"
	fieldSelector, err := fields.ParseSelector(fmt.Sprintf(selFmt, node, v1.PodSucceeded, v1.PodFailed))
	if err != nil {
		return nil, err
	}

	return a.DialOrDie().Core().Pods("").List(metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
}

// IsNamespaced check on server if given resource is namespaced
func (a *APIClient) IsNamespaced(res string) bool {
	list, _ := a.DialOrDie().Discovery().ServerPreferredResources()
	for _, l := range list {
		log.Debug().Msgf("GV %s", l.GroupVersion)
		for _, r := range l.APIResources {
			if r.Name == res {
				return r.Namespaced
			}
		}
	}
	return false
}

// SupportsResource checks for resource supported version against the server.
func (a *APIClient) SupportsResource(group string) bool {
	list, err := a.DialOrDie().Discovery().ServerPreferredResources()
	if err != nil {
		log.Debug().Err(err).Msg("Unable to dial api server")
		return false
	}
	for _, l := range list {
		log.Debug().Msgf(">>> Group %s", l.GroupVersion)
		if l.GroupVersion == group {
			return true
		}
	}
	return false
}

// Config return a kubernetes configuration.
func (a *APIClient) Config() *Config {
	return a.config
}

// HasMetrics returns true if the cluster supports metrics.
func (a *APIClient) HasMetrics() bool {
	return a.useMetricServer
}

// DialOrDie returns a handle to api server or die.
func (a *APIClient) DialOrDie() kubernetes.Interface {
	if a.client != nil {
		return a.client
	}

	var err error
	if a.client, err = kubernetes.NewForConfig(a.RestConfigOrDie()); err != nil {
		a.log.Fatal().Msgf("Unable to connect to api server %v", err)
	}
	return a.client
}

// RestConfigOrDie returns a rest api client.
func (a *APIClient) RestConfigOrDie() *restclient.Config {
	cfg, err := a.config.RESTConfig()
	if err != nil {
		a.log.Panic().Msgf("Unable to connect to api server %v", err)
	}
	return cfg
}

// DynDialOrDie returns a handle to a dynamic interface.
func (a *APIClient) DynDialOrDie() dynamic.Interface {
	if a.dClient != nil {
		return a.dClient
	}

	var err error
	if a.dClient, err = dynamic.NewForConfig(a.RestConfigOrDie()); err != nil {
		a.log.Panic().Err(err)
	}
	return a.dClient
}

// NSDialOrDie returns a handle to a namespaced resource.
func (a *APIClient) NSDialOrDie() dynamic.NamespaceableResourceInterface {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.nsClient != nil {
		return a.nsClient
	}
	a.nsClient = a.DynDialOrDie().Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	})

	return a.nsClient
}

// MXDial returns a handle to the metrics server.
func (a *APIClient) MXDial() (*versioned.Clientset, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.mxsClient != nil {
		return a.mxsClient, nil
	}
	var err error
	if a.mxsClient, err = versioned.NewForConfig(a.RestConfigOrDie()); err != nil {
		a.log.Debug().Err(err)
	}

	return a.mxsClient, err
}

// SwitchContextOrDie handles kubeconfig context switches.
func (a *APIClient) SwitchContextOrDie(ctx string) {
	currentCtx, err := a.config.CurrentContextName()
	if err != nil {
		panic(err)
	}

	if currentCtx != ctx {
		a.reset()
		if err := a.config.SwitchContext(ctx); err != nil {
			panic(err)
		}
		a.useMetricServer = a.supportsMxServer()
	}
}

func (a *APIClient) reset() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.client, a.dClient, a.nsClient, a.mxsClient = nil, nil, nil, nil
}

func (a *APIClient) supportsMxServer() bool {
	apiGroups, err := a.DialOrDie().Discovery().ServerGroups()
	if err != nil {
		return false
	}

	for _, discoveredAPIGroup := range apiGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}

		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}

	return false
}

// SupportsRes checks latest supported version.
func (a *APIClient) SupportsRes(group string, versions []string) (string, bool, error) {
	apiGroups, err := a.DialOrDie().Discovery().ServerGroups()
	if err != nil {
		return "", false, err
	}

	for _, grp := range apiGroups.Groups {
		if grp.Name != group {
			continue
		}
		return grp.PreferredVersion.Version, true, nil
	}

	return "", false, nil
}
