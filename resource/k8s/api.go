package k8s

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/pkg/kubectl/metricsutil"
	"k8s.io/kubernetes/staging/src/k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NA Not available
const NA = "n/a"

var (
	conn                        connection
	supportedMetricsAPIVersions = []string{"v1beta1"}
)

func init() {
	conn = &apiServer{}
}

type (
	// ApiGroup represents a K8s resource descriptor.
	ApiGroup struct {
		Resource             string
		Group, Kind, Version string
		Plural, Singular     string
		Aliases              []string
	}

	// Collection of empty interfaces.
	Collection []interface{}

	// Res K8s api server calls
	Res interface {
		Get(ns string, name string) (interface{}, error)
		List(ns string) (Collection, error)
		Delete(ns string, name string) error
	}

	connection interface {
		restConfigOrDie() *restclient.Config
		apiConfigOrDie() *clientcmdapi.Config
		dialOrDie() kubernetes.Interface
		dynDialOrDie() dynamic.Interface
		nsDialOrDie() dynamic.NamespaceableResourceInterface
		mxsDial() (*versioned.Clientset, error)
		heapsterDial() (*metricsutil.HeapsterMetricsClient, error)
		getClusterName() string
		setClusterName(n string)
		hasMetricsServer() bool
	}

	apiServer struct {
		client          kubernetes.Interface
		dClient         dynamic.Interface
		csClient        *clientset.Clientset
		nsClient        dynamic.NamespaceableResourceInterface
		heapsterClient  *metricsutil.HeapsterMetricsClient
		mxsClient       *versioned.Clientset
		cluster         string
		useMetricServer bool
	}
)

func (a *apiServer) getClusterName() string {
	return a.cluster
}

func (a *apiServer) setClusterName(n string) {
	a.cluster = n
}

func (a *apiServer) hasMetricsServer() bool {
	return a.useMetricServer
}

// DialOrDie returns a handle to api server or die.
func (a *apiServer) dialOrDie() kubernetes.Interface {
	a.checkCurrentConfig()
	if a.client != nil {
		return a.client
	}

	var err error
	if a.client, err = kubernetes.NewForConfig(a.restConfigOrDie()); err != nil {
		panic(err)
	}

	return a.client
}

func (a *apiServer) csDialOrDie() *clientset.Clientset {
	a.checkCurrentConfig()
	if a.csClient != nil {
		return a.csClient
	}

	var cfg *rest.Config
	// cfg := clientcmd.NewNonInteractiveClientConfig(config, contextName, overrides, configAccess)
	var err error
	if a.csClient, err = clientset.NewForConfig(cfg); err != nil {
		panic(err)
	}

	return a.csClient
}

// DynDial returns a handle to the api server.
func (a *apiServer) dynDialOrDie() dynamic.Interface {
	a.checkCurrentConfig()
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
	a.checkCurrentConfig()
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
	a.checkCurrentConfig()
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
	a.checkCurrentConfig()
	if a.mxsClient != nil {
		return a.mxsClient, nil
	}

	opts := clientcmd.NewDefaultPathOptions()
	dat, err := ioutil.ReadFile(opts.GetDefaultFilename())
	if err != nil {
		return nil, err
	}
	rc, err := clientcmd.RESTConfigFromKubeConfig(dat)
	if err != nil {
		return nil, err
	}

	a.mxsClient, err = versioned.NewForConfig(rc)
	return a.mxsClient, err
}

// ConfigOrDie check kubernetes cluster config.
// Dies if no config is found or incorrect.
func ConfigOrDie() {
	var srv *apiServer
	cfg := srv.apiConfigOrDie()
	if clientcmdapi.IsConfigEmpty(cfg) {
		panic("K8s config file load failed. Please check your .kube/config or $KUBECONFIG env")
	}
}

func (*apiServer) apiConfigOrDie() *clientcmdapi.Config {
	paths := clientcmd.NewDefaultPathOptions()
	c, err := paths.GetStartingConfig()
	if err != nil {
		panic(err)
	}
	return c
}

func (*apiServer) restConfigOrDie() *restclient.Config {
	opts := clientcmd.NewDefaultPathOptions()
	cfg, err := clientcmd.BuildConfigFromFlags("", opts.GetDefaultFilename())
	if err != nil {
		panic(err)
	}
	return cfg
}

func (a *apiServer) checkCurrentConfig() {
	cfg, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		log.Fatal(err)
	}

	if len(a.getClusterName()) == 0 {
		a.setClusterName(cfg.CurrentContext)
		a.useMetricServer = a.supportsMxServer()
		return
	}

	if a.getClusterName() != cfg.CurrentContext {
		a.reset()
		a.setClusterName(cfg.CurrentContext)
		a.useMetricServer = a.supportsMxServer()
	}
}

func (a *apiServer) reset() {
	a.client, a.dClient, a.nsClient = nil, nil, nil
	a.heapsterClient, a.mxsClient = nil, nil
	a.setClusterName("")
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
