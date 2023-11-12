package dao

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/port"
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

// PortForwarder tracks a port forward stream.
type PortForwarder struct {
	Factory
	genericclioptions.IOStreams

	stopChan, readyChan chan struct{}
	active              bool
	path                string
	tunnel              port.PortTunnel
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

// String dumps as string.
func (p *PortForwarder) String() string {
	return fmt.Sprintf("%s|%s", p.path, p.tunnel)
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

// Port returns the port mapping.
func (p *PortForwarder) Port() string {
	return p.tunnel.PortMap()
}

// ContainerPort returns the container port.
func (p *PortForwarder) ContainerPort() string {
	return p.tunnel.ContainerPort
}

// LocalPort returns the local port.
func (p *PortForwarder) LocalPort() string {
	return p.tunnel.LocalPort
}

// ID returns a pf id.
func (p *PortForwarder) ID() string {
	return PortForwardID(p.path, p.tunnel.Container, p.tunnel.PortMap())
}

// Container returns the target's container.
func (p *PortForwarder) Container() string {
	return p.tunnel.Container
}

// Stop terminates a port forward.
func (p *PortForwarder) Stop() {
	log.Debug().Msgf("<<< Stopping PortForward %s", p.ID())
	p.active = false
	close(p.stopChan)
}

// FQN returns the portforward unique id.
func (p *PortForwarder) FQN() string {
	return p.path + ":" + p.tunnel.Container
}

// HasPortMapping checks if port mapping is defined for this fwd.
func (p *PortForwarder) HasPortMapping(portMap string) bool {
	return p.tunnel.PortMap() == portMap
}

// Start initiates a port forward session for a given pod and ports.
func (p *PortForwarder) Start(path string, tt port.PortTunnel) (*portforward.PortForwarder, error) {
	p.path, p.tunnel, p.age = path, tt, time.Now()

	ns, n := client.Namespaced(path)
	auth, err := p.Client().CanI(ns, "v1/pods", []string{client.GetVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to get pods")
	}

	podName := strings.Split(n, "|")[0]
	var res Pod
	res.Init(p, client.NewGVR("v1/pods"))
	pod, err := res.GetInstance(client.FQN(ns, podName))
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
		Name(podName).
		SubResource("portforward")

	return p.forwardPorts("POST", req.URL(), tt.Address, tt.PortMap())
}

func (p *PortForwarder) forwardPorts(method string, url *url.URL, addr, portMap string) (*portforward.PortForwarder, error) {
	cfg, err := p.Client().Config().RESTConfig()
	if err != nil {
		return nil, err
	}
	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)

	return portforward.NewOnAddresses(dialer, []string{addr}, []string{portMap}, p.stopChan, p.readyChan, p.Out, p.ErrOut)
}

// ----------------------------------------------------------------------------
// Helpers...

// PortForwardID computes port-forward identifier.
func PortForwardID(path, co, portMap string) string {
	if strings.Contains(path, "|") {
		return path + "|" + portMap
	}

	return path + "|" + co + "|" + portMap
}

func codec() (serializer.CodecFactory, runtime.ParameterCodec) {
	scheme := runtime.NewScheme()
	gv := schema.GroupVersion{Group: "", Version: "v1"}
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypes(gv, &metav1beta1.Table{}, &metav1beta1.TableOptions{})
	scheme.AddKnownTypes(metav1beta1.SchemeGroupVersion, &metav1beta1.Table{}, &metav1beta1.TableOptions{})

	return serializer.NewCodecFactory(scheme), runtime.NewParameterCodec(scheme)
}
