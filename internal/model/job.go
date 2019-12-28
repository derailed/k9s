package model

import (
	"context"
	"errors"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Job represents a collections of jobs.
type Job struct {
	Resource
}

// List returns a collection of screen dumps.
func (c *Job) List(ctx context.Context) ([]runtime.Object, error) {
	uid, ok := ctx.Value(internal.KeyUID).(string)
	if !ok {
		log.Debug().Msgf("NO UID in context")
	}
	path, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("no cronjob path found in context")
	}

	log.Debug().Msgf("Listing jobs %q %q--%q", c.gvr, uid, path)
	oo, err := c.Resource.List(ctx)
	if err != nil {
		return nil, err
	}
	if uid == "" {
		return oo, nil
	}

	_, cronName := client.Namespaced(path)
	jj := make([]runtime.Object, 0, len(oo))
	for _, j := range oo {
		var job batchv1.Job
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(j.(*unstructured.Unstructured).Object, &job)
		if err != nil {
			return nil, err
		}
		log.Debug().Msgf("Looking at job %q -- %q", job.Name, cronName)
		if !isNamedAfter(cronName, job.Name) {
			continue
		}
		log.Debug().Msgf("GOT Job %s -- %#v -- %q -- %q", job.Name, job.Labels, uid, path)
		jj = append(jj, j)
	}

	return jj, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func isControlledBy(cuid, id string) bool {
	tokens := strings.Split(cuid, "-")
	root := strings.Join(tokens[2:], "-")
	return strings.Contains(id, root)
}

func isNamedAfter(p, n string) bool {
	tokens := strings.Split(n, "-")
	if len(tokens) == 0 || tokens[0] != p {
		return false
	}
	return true
}
