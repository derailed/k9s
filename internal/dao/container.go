package dao

import (
	"context"
	"errors"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
)

type Container struct {
	Generic
}

var _ Accessor = &Container{}
var _ Loggable = &Container{}

// Logs tails a given container logs
func (c *Container) TailLogs(ctx context.Context, logChan chan<- string, opts LogOptions) error {
	log.Debug().Msgf("CO TAILLOGS %#v", ctx)
	log.Debug().Msgf("CO TAILLOGS %#v", opts)

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

	return tailLogs(ctx, c, logChan, opts)
}

// Logs fetch container logs for a given pod and container.
func (c *Container) Logs(path string, opts *v1.PodLogOptions) *restclient.Request {
	ns, n := client.Namespaced(path)
	return c.Client().DialOrDie().CoreV1().Pods(ns).GetLogs(n, opts)
}
