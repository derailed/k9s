package k8s

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
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

// PortForward tracks a port forward stream.
type PortForward struct {
	Connection
	genericclioptions.IOStreams

	stopChan, readyChan chan struct{}
	logger              *zerolog.Logger
	active              bool
	path                string
	container           string
	ports               []string
	age                 time.Time
}

// NewPortForward returns a new port forward streamer.
func NewPortForward(c Connection, l *zerolog.Logger) *PortForward {
	return &PortForward{
		Connection: c,
		logger:     l,
		stopChan:   make(chan struct{}),
		readyChan:  make(chan struct{}),
	}
}

// Age returns the port forward age.
func (p *PortForward) Age() string {
	return time.Since(p.age).String()
}

// Active returns the forward status.
func (p *PortForward) Active() bool {
	return p.active
}

// SetActive mark a portforward as active.
func (p *PortForward) SetActive(b bool) {
	p.active = b
}

// Ports returns the forwarded ports mappings.
func (p *PortForward) Ports() []string {
	return p.ports
}

// Path returns the pod resource path.
func (p *PortForward) Path() string {
	return p.path
}

// Container returns the targetes container.
func (p *PortForward) Container() string {
	return p.container
}

// Stop terminates a port forard
func (p *PortForward) Stop() {
	p.logger.Debug().Msgf("<<< Stopping port forward %q %v", p.path, p.ports)
	p.active = false
	close(p.stopChan)
}

// FQN returns the portforward unique id.
func (p *PortForward) FQN() string {
	return p.path + ":" + p.container
}

// Start initiates a port forward session for a given pod and ports.
func (p *PortForward) Start(path, co string, ports []string) (*portforward.PortForwarder, error) {
	p.path, p.container, p.ports, p.age = path, co, ports, time.Now()

	ns, n := namespaced(path)
	pod, err := p.DialOrDie().CoreV1().Pods(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if pod.Status.Phase != v1.PodRunning {
		return nil, fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	rcfg := p.RestConfigOrDie()
	rcfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	rcfg.APIPath = "/api"
	codec, _ := codec()
	rcfg.NegotiatedSerializer = codec.WithoutConversion()
	clt, err := rest.RESTClientFor(rcfg)
	if err != nil {
		p.logger.Debug().Msgf("Boom! %#v", err)
		return nil, err
	}
	req := clt.Post().
		Resource("pods").
		Namespace(ns).
		Name(n).
		SubResource("portforward")

	return p.forwardPorts("POST", req.URL(), ports)
}

func (p *PortForward) forwardPorts(method string, url *url.URL, ports []string) (*portforward.PortForwarder, error) {
	cfg, err := p.Config().RESTConfig()
	if err != nil {
		return nil, err
	}
	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	addrs := []string{localhost}
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
