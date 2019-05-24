package resource

import (
	"bufio"
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// Container represents a container on a pod.
	Container struct {
		*Base

		pod           *v1.Pod
		isInit        bool
		instance      v1.Container
		MetricsServer MetricsServer
		metrics       *mv1beta1.PodMetrics
		mx            sync.RWMutex
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

// SetPodMetrics set the current k8s resource metrics on associated pod.
func (r *Container) SetPodMetrics(m *mv1beta1.PodMetrics) {
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
			var closes bool
			r.mx.RLock()
			{
				closes = blocked
			}
			r.mx.RUnlock()
			if closes {
				log.Debug().Msg(">>Closing Channel<<")
				close(c)
				cancel()
			}
		}
	}()

	// This call will block if nothing is in the stream!!
	stream, err := req.Stream()
	if err != nil {
		log.Warn().Err(err).Msgf("Stream canceled `%s/%s:%s", ns, n, co)
		return cancel, err
	}

	r.mx.Lock()
	{
		blocked = false
	}
	r.mx.Unlock()

	go func() {
		defer func() {
			log.Debug().Msgf("Closing stream %s:%s", n, co)
			close(c)
			stream.Close()
			cancel()
		}()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			c <- scanner.Text()
			select {
			case <-ctx.Done():
				return
			default:
			}
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
	return append(Row{},
		"NAME",
		"IMAGE",
		"READY",
		"STATE",
		"RS",
		"PROBES(L:R)",
		"CPU",
		"MEM",
		"%CPU",
		"%MEM",
		"AGE",
	)
}

// Fields retrieves displayable fields.
func (r *Container) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	scpu, smem, pcpu, pmem := NAValue, NAValue, NAValue, NAValue
	if r.metrics != nil {
		var (
			cpu int64
			mem float64
		)
		for _, co := range r.metrics.Containers {
			if co.Name == i.Name {
				cpu = co.Usage.Cpu().MilliValue()
				mem = k8s.ToMB(co.Usage.Memory().Value())
				break
			}
		}
		scpu, smem = ToMillicore(cpu), ToMi(mem)
		rcpu, rmem := containerResources(i)
		if rcpu != nil {
			pcpu = AsPerc(toPerc(float64(cpu), float64(rcpu.MilliValue())))
		}
		if rmem != nil {
			pmem = AsPerc(toPerc(mem, k8s.ToMB(rmem.Value())))
		}
	}

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

	ready, state, restarts := "false", MissingValue, "0"
	if cs != nil {
		ready, state, restarts = boolToStr(cs.Ready), toState(cs.State), strconv.Itoa(int(cs.RestartCount))
	}

	return append(ff,
		i.Name,
		i.Image,
		ready,
		state,
		restarts,
		probe(i.LivenessProbe)+":"+probe(i.ReadinessProbe),
		scpu,
		smem,
		pcpu,
		pmem,
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

func probe(p *v1.Probe) string {
	if p == nil {
		return "on"
	}
	return "off"
}

func asMi(v int64) float64 {
	return float64(v) / 1024 * 1024
}
