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
	Factory
	genericclioptions.IOStreams

	stopChan, readyChan chan struct{}
	active              bool
	path                string
	container           string
	ports               []string
	age                 time.Time
}

// NewPortForwarder returns a new port forward streamer.
func NewPortForwarder(f Factory) *PortForwarder {
	return &PortForwarder{
		Factory:   f,
		stopChan:  make(chan struct{}),
		readyChan: make(chan struct{}),
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
	return PortForwardID(p.path, p.container)
}

// PortForwardID computes port-forward identifier.
func PortForwardID(path, co string) string {
	return path + ":" + co
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

// HasPortMapping checks if port mapping is defined for this fwd.
func (p *PortForwarder) HasPortMapping(m string) bool {
	for _, mapping := range p.ports {
		if mapping == m {
			return true
		}
	}
	return false
}

// Start initiates a port forward session for a given pod and ports.
func (p *PortForwarder) Start(path, co string, tt []client.PortTunnel) (*portforward.PortForwarder, error) {
	if len(tt) == 0 {
		return nil, fmt.Errorf("no ports assigned")
	}
	fwds := make([]string, 0, len(tt))
	for _, t := range tt {
		fwds = append(fwds, t.PortMap())
	}
	p.path, p.container, p.ports, p.age = path, co, fwds, time.Now()

	ns, n := client.Namespaced(path)
	auth, err := p.Client().CanI(ns, "v1/pods", []string{client.GetVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to get pods")
	}

	var res Pod
	res.Init(p, client.NewGVR("v1/pods"))
	pod, err := res.GetInstance(path)
	if err != nil {
		return nil, err
	}
	if pod.Status.Phase != v1.PodRunning {
		return nil, fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	auth, err = p.Client().CanI(ns, "v1/pods:portforward", []string{client.CreateVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to update portforward")
	}

	cfg, err := p.Client().RestConfig()
	if err != nil {
		return nil, err
	}
	cfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	cfg.APIPath = "/api"
	codec, _ := codec()
	cfg.NegotiatedSerializer = codec.WithoutConversion()
	clt, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}
	req := clt.Post().
		Resource("pods").
		Namespace(ns).
		Name(n).
		SubResource("portforward")

	return p.forwardPorts("POST", req.URL(), tt[0].Address, fwds)
}

func (p *PortForwarder) forwardPorts(method string, url *url.URL, address string, ports []string) (*portforward.PortForwarder, error) {
	cfg, err := p.Client().Config().RESTConfig()
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
