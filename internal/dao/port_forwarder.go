package dao

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const localhost = "localhost"

// PortForwarder tracks a port forward stream.
type PortForwarder struct {
	client.Connection
	genericclioptions.IOStreams

	stopChan, readyChan chan struct{}
	active              bool
	path                string
	container           string
	ports               []string
	age                 time.Time
}

// NewPortForwarder returns a new port forward streamer.
func NewPortForwarder(c client.Connection) *PortForwarder {
	return &PortForwarder{
		Connection: c,
		stopChan:   make(chan struct{}),
		readyChan:  make(chan struct{}),
	}
}

// Age returns the port forward age.
func (p *PortForwarder) Age() string {
	return time.Since(p.age).String()
}

// Active returns the forward status.
func (p *PortForwarder) Active() bool {
	return p.active
}

// SetActive mark a portforward as active.
func (p *PortForwarder) SetActive(b bool) {
	p.active = b
}

// Ports returns the forwarded ports mappings.
func (p *PortForwarder) Ports() []string {
	return p.ports
}

// Path returns the pod resource path.
func (p *PortForwarder) Path() string {
	return p.path + ":" + p.container
}

// Container returns the targetes container.
func (p *PortForwarder) Container() string {
	return p.container
}

// Stop terminates a port forard
func (p *PortForwarder) Stop() {
	log.Debug().Msgf("<<< Stopping PortForward %q %v", p.path, p.ports)
	p.active = false
	close(p.stopChan)
}

// FQN returns the portforward unique id.
func (p *PortForwarder) FQN() string {
	return p.path + ":" + p.container
}

// Start initiates a port forward session for a given pod and ports.
func (p *PortForwarder) Start(path, co, address string, ports []string) (*portforward.PortForwarder, error) {
	p.path, p.container, p.ports, p.age = path, co, ports, time.Now()

	ns, n := client.Namespaced(path)
	auth, err := p.CanI(ns, "v1/pods", []string{client.GetVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to get pods")
	}
	pod, err := p.DialOrDie().CoreV1().Pods(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if pod.Status.Phase != v1.PodRunning {
		return nil, fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	auth, err = p.CanI(ns, "v1/pods:portforward", []string{client.UpdateVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to update portforward")
	}

	rcfg := p.RestConfigOrDie()
	rcfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	rcfg.APIPath = "/api"
	codec, _ := codec()
	rcfg.NegotiatedSerializer = codec.WithoutConversion()
	clt, err := rest.RESTClientFor(rcfg)
	if err != nil {
		log.Debug().Msgf("Boom! %#v", err)
		return nil, err
	}
	req := clt.Post().
		Resource("pods").
		Namespace(ns).
		Name(n).
		SubResource("portforward")

	return p.forwardPorts("POST", req.URL(), address, ports)
}

func (p *PortForwarder) forwardPorts(method string, url *url.URL, address string, ports []string) (*portforward.PortForwarder, error) {
	cfg, err := p.Config().RESTConfig()
	if err != nil {
		return nil, err
	}
	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	if address == "" {
		address = localhost
	}
	addrs := strings.Split(address, ",")
	return portforward.NewOnAddresses(dialer, addrs, ports, p.stopChan, p.readyChan, p.Out, p.ErrOut)
}

// ----------------------------------------------------------------------------
// Helpers...

func codec() (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	gv := schema.GroupVersion{Group: "", Version: "v1"}
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})

	return serializer.NewCodecFactory(scheme), runtime.NewParameterCodec(scheme)
}
