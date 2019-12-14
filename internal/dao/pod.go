package dao

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
)

const defaultTimeout = 1 * time.Second

// Pod represents a pod resource.
type Pod struct {
	Generic
}

var _ Accessor = &Pod{}
var _Loggable = &Pod{}

// Logs fetch container logs for a given pod and container.
func (p *Pod) Logs(path string, opts *v1.PodLogOptions) *restclient.Request {
	ns, n := k8s.Namespaced(path)
	return p.Client().DialOrDie().CoreV1().Pods(ns).GetLogs(n, opts)
}

// Containers returns all container names on pod
func (p *Pod) Containers(path string, includeInit bool) ([]string, error) {
	o, err := p.Get("v1/pod", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var pod v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
	if err != nil {
		return nil, err
	}

	cc := []string{}
	for _, c := range pod.Spec.Containers {
		cc = append(cc, c.Name)
	}

	if includeInit {
		for _, c := range pod.Spec.InitContainers {
			cc = append(cc, c.Name)
		}
	}

	return cc, nil
}

// Logs tails a given container logs
func (p *Pod) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	if !opts.HasContainer() {
		return p.logs(ctx, c, opts)
	}
	return tailLogs(ctx, p, c, opts)
}

// PodLogs tail logs for all containers in a running Pod.
func (p *Pod) logs(ctx context.Context, c chan<- string, opts LogOptions) error {
	fac, ok := ctx.Value(internal.KeyFactory).(*watch.Factory)
	if !ok {
		return errors.New("Expecting an informer")
	}
	o, err := fac.Get("v1/pods", opts.Path, labels.Everything())
	if err != nil {
		return err
	}

	var po v1.Pod
	if runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po); err != nil {
		return err
	}
	opts.Color = asColor(po.Name)
	if len(po.Spec.InitContainers)+len(po.Spec.Containers) == 1 {
		opts.SingleContainer = true
	}

	for _, co := range po.Spec.InitContainers {
		opts.Container = co.Name
		if err := p.TailLogs(ctx, c, opts); err != nil {
			return err
		}
	}
	rcos := loggableContainers(po.Status)
	for _, co := range po.Spec.Containers {
		if in(rcos, co.Name) {
			opts.Container = co.Name
			if err := p.TailLogs(ctx, c, opts); err != nil {
				log.Error().Err(err).Msgf("Getting logs for %s failed", co.Name)
				return err
			}
		}
	}

	return nil
}

func tailLogs(ctx context.Context, logger Logger, c chan<- string, opts LogOptions) error {
	log.Debug().Msgf("Tailing logs for %q -- %q", opts.Path, opts.Container)
	o := v1.PodLogOptions{
		Container: opts.Container,
		Follow:    true,
		TailLines: &opts.Lines,
		Previous:  opts.Previous,
	}
	req := logger.Logs(opts.Path, &o)
	ctxt, cancelFunc := context.WithCancel(ctx)
	req.Context(ctxt)

	var blocked int32 = 1
	go logsTimeout(cancelFunc, &blocked)

	// This call will block if nothing is in the stream!!
	stream, err := req.Stream()
	atomic.StoreInt32(&blocked, 0)
	if err != nil {
		log.Error().Err(err).Msgf("Log stream failed for `%s", opts.Path)
		return fmt.Errorf("Unable to obtain log stream for %s", opts.Path)
	}
	go readLogs(ctx, stream, c, opts)

	return nil
}

func logsTimeout(cancel context.CancelFunc, blocked *int32) {
	<-time.After(defaultTimeout)
	if atomic.LoadInt32(blocked) == 1 {
		log.Debug().Msg("Timed out reading the log stream")
		cancel()
	}
}

func readLogs(ctx context.Context, stream io.ReadCloser, c chan<- string, opts LogOptions) {
	defer func() {
		log.Debug().Msgf(">>> Closing stream `%s", opts.Path)
		if err := stream.Close(); err != nil {
			log.Error().Err(err).Msg("Cloing stream")
		}
	}()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			c <- opts.DecorateLog(scanner.Text())
		}
	}
}

// Helpers...

func loggableContainers(s v1.PodStatus) []string {
	var rcos []string
	for _, c := range s.ContainerStatuses {
		rcos = append(rcos, c.Name)
	}
	return rcos
}

func asColor(n string) color.Paint {
	var sum int
	for _, r := range n {
		sum += int(r)
	}
	return color.Paint(30 + 2 + sum%6)
}

// Check if string is in a string list.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}
	return false
}
