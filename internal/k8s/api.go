package k8s

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/pkg/kubectl/metricsutil"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NA Not available
const NA = "n/a"

var (
	conn                        = &apiServer{}
	supportedMetricsAPIVersions = []string{"v1beta1"}
)

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

	// Res K8s api server calls.
	Res interface {
		Get(ns string, name string) (interface{}, error)
		List(ns string) (Collection, error)
		Delete(ns string, name string) error
	}

	// Connection represents a k8s api server connection.
	connection interface {
		configAccess() clientcmd.ConfigAccess
		restConfigOrDie() *restclient.Config
		apiConfigOrDie() clientcmdapi.Config
		dialOrDie() kubernetes.Interface
		dynDialOrDie() dynamic.Interface
		nsDialOrDie() dynamic.NamespaceableResourceInterface
		mxsDial() (*versioned.Clientset, error)
		heapsterDial() (*metricsutil.HeapsterMetricsClient, error)
		hasMetricsServer() bool
	}

	apiServer struct {
		config          *Config
		client          kubernetes.Interface
		dClient         dynamic.Interface
		nsClient        dynamic.NamespaceableResourceInterface
		heapsterClient  *metricsutil.HeapsterMetricsClient
		mxsClient       *versioned.Clientset
		useMetricServer bool
	}
)

// InitConnectionOrDie initialize connection from command line args.
// Checks for connectivity with the api server.
func InitConnectionOrDie(config *Config) {
	conn = &apiServer{config: config}
	conn.useMetricServer = conn.supportsMxServer()
}

func (a *apiServer) hasMetricsServer() bool {
	return a.useMetricServer
}

// DialOrDie returns a handle to api server or die.
func (a *apiServer) dialOrDie() kubernetes.Interface {
	if a.client != nil {
		return a.client
	}

	var err error
	if a.client, err = kubernetes.NewForConfig(a.restConfigOrDie()); err != nil {
		panic(err)
	}
	return a.client
}

// DynDial returns a handle to the api server.
func (a *apiServer) dynDialOrDie() dynamic.Interface {
	if a.dClient != nil {
		return a.dClient
	}

	var err error
	if a.dClient, err = dynamic.NewForConfig(a.restConfigOrDie()); err != nil {
		panic(err)
	}

	return a.dClient
}

func (a *apiServer) nsDialOrDie() dynamic.NamespaceableResourceInterface {
	if a.nsClient != nil {
		return a.nsClient
	}

	a.nsClient = a.dynDialOrDie().Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	})
	return a.nsClient
}

func (a *apiServer) heapsterDial() (*metricsutil.HeapsterMetricsClient, error) {
	if a.heapsterClient != nil {
		return a.heapsterClient, nil
	}

	a.heapsterClient = metricsutil.NewHeapsterMetricsClient(
		a.dialOrDie().CoreV1(),
		metricsutil.DefaultHeapsterNamespace,
		metricsutil.DefaultHeapsterScheme,
		metricsutil.DefaultHeapsterService,
		metricsutil.DefaultHeapsterPort,
	)
	return a.heapsterClient, nil
}

func (a *apiServer) mxsDial() (*versioned.Clientset, error) {
	if a.mxsClient != nil {
		return a.mxsClient, nil
	}

	var err error
	a.mxsClient, err = versioned.NewForConfig(a.restConfigOrDie())
	return a.mxsClient, err
}

func (a *apiServer) restConfigOrDie() *restclient.Config {
	cfg, err := a.config.RESTConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

func (a *apiServer) switchContextOrDie(ctx string) {
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

func (a *apiServer) reset() {
	a.client, a.dClient, a.nsClient = nil, nil, nil
	a.heapsterClient, a.mxsClient = nil, nil
}

func (a *apiServer) supportsMxServer() bool {
	apiGroups, err := a.dialOrDie().Discovery().ServerGroups()
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
