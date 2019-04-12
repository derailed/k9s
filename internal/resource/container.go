package resource

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

type (
	// Container represents a container on a pod.
	Container struct {
		*Base

		pod           *v1.Pod
		isInit        bool
		instance      v1.Container
		MetricsServer MetricsServer
		metrics       k8s.PodMetrics
	}
)

// NewContainerList returns a collection of container.
func NewContainerList(c Connection, mx MetricsServer, pod *v1.Pod) List {
	return NewList(
		"",
		"co",
		NewContainer(c, mx, pod),
		0,
	)
}

// NewContainer returns a new set of containers.
func NewContainer(c Connection, mx MetricsServer, pod *v1.Pod) *Container {
	co := Container{
		Base:          &Base{Connection: c, Resource: k8s.NewPod(c)},
		pod:           pod,
		MetricsServer: mx,
		metrics:       k8s.PodMetrics{},
	}
	co.Factory = &co

	return &co
}

// New builds a new Container instance from a k8s resource.
func (r *Container) New(i interface{}) Columnar {
	co := NewContainer(r.Connection, r.MetricsServer, r.pod)
	co.instance = i.(v1.Container)
	co.path = r.namespacedName(r.pod.ObjectMeta) + ":" + co.instance.Name

	return co
}

// Metrics retrieves cpu/mem resource consumption on associated pod.
func (r *Container) Metrics() k8s.PodMetrics {
	return r.metrics
}

// SetMetrics set the current k8s resource metrics on associated pod.
func (r *Container) SetMetrics(m k8s.PodMetrics) {
	r.metrics = m
}

// Marshal resource to yaml.
func (r *Container) Marshal(path string) (string, error) {
	return "", nil
}

// Logs tails a given container logs
func (r *Container) Logs(c chan<- string, ns, n, co string, lines int64, prev bool) (context.CancelFunc, error) {
	req := r.Resource.(k8s.Loggable).Logs(ns, n, co, lines, prev)
	ctx, cancel := context.WithCancel(context.TODO())
	req.Context(ctx)

	blocked := true
	go func() {
		select {
		case <-time.After(defaultTimeout):
			if blocked {
				close(c)
				cancel()
			}
		}
	}()
	// This call will block if nothing is in the stream!!
	stream, err := req.Stream()
	blocked = false
	if err != nil {
		log.Error().Msgf("Tail logs failed `%s/%s:%s -- %v", ns, n, co, err)
		return cancel, fmt.Errorf("%v", err)
	}

	go func() {
		defer func() {
			stream.Close()
			cancel()
			close(c)
		}()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			c <- scanner.Text()
		}
	}()

	return cancel, nil
}

// List resources for a given namespace.
func (r *Container) List(ns string) (Columnars, error) {
	icos := r.pod.Spec.InitContainers
	cos := r.pod.Spec.Containers

	cc := make(Columnars, 0, len(icos)+len(cos))
	for _, co := range icos {
		ci := r.New(co)
		ci.(*Container).isInit = true
		cc = append(cc, ci)
	}
	for _, co := range cos {
		cc = append(cc, r.New(co))
	}

	return cc, nil
}

// Header return resource header.
func (*Container) Header(ns string) Row {
	hh := Row{}

	return append(hh,
		"NAME",
		"IMAGE",
		"READY",
		"STATE",
		"RS",
		"LPROB",
		"RPROB",
		"CPU",
		"MEM",
		"RCPU",
		"RMEM",
		"AGE",
	)
}

// Fields retrieves displayable fields.
func (r *Container) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	mxs, _ := r.MetricsServer.FetchPodsMetrics(r.pod.Namespace)

	var cpu, mem string
	for _, mx := range mxs {
		if mx.Name != r.pod.Name {
			continue
		}
		for _, co := range mx.Containers {
			if co.Name != i.Name {
				continue
			}
			cpu, mem = toRes(co.Usage)
		}
	}

	rcpu, rmem := resources(i)

	var cs *v1.ContainerStatus
	for _, c := range r.pod.Status.ContainerStatuses {
		if c.Name != i.Name {
			continue
		}
		cs = &c
	}
	if cs == nil {
		for _, c := range r.pod.Status.InitContainerStatuses {
			if c.Name != i.Name {
				continue
			}
			cs = &c
		}
	}

	return append(ff,
		i.Name,
		i.Image,
		boolToStr(cs.Ready),
		toState(cs.State),
		strconv.Itoa(int(cs.RestartCount)),
		probe(i.LivenessProbe),
		probe(i.ReadinessProbe),
		cpu,
		mem,
		rcpu,
		rmem,
		toAge(r.pod.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func toState(s v1.ContainerState) string {
	switch {
	case s.Waiting != nil:
		if s.Waiting.Reason != "" {
			return s.Waiting.Reason
		}
		return "Waiting"

	case s.Terminated != nil:
		if s.Terminated.Reason != "" {
			return s.Terminated.Reason
		}
		return "Terminated"
	case s.Running != nil:
		return "Running"
	default:
		return MissingValue
	}
}

func toRes(r v1.ResourceList) (string, string) {
	cpu, mem := r[v1.ResourceCPU], r[v1.ResourceMemory]

	return ToMillicore(cpu.MilliValue()), ToMi(k8s.ToMB(mem.Value()))
}

func resources(c v1.Container) (cpu, mem string) {
	req, lim := c.Resources.Requests, c.Resources.Limits

	if len(req) == 0 {
		if len(lim) != 0 {
			return toRes(lim)
		}
	} else {
		return toRes(req)
	}

	return "0", "0"
}

func probe(p *v1.Probe) string {
	if p == nil {
		return "no"
	}

	return "yes"
}

func asMi(v int64) float64 {
	const megaByte = 1024 * 1024

	return float64(v) / megaByte
}
