// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

var (
	_ ImageLister = (*ReplicaSet)(nil)
)

// ReplicaSet represents a replicaset K8s resource.
type ReplicaSet struct {
	Resource
}

// ListImages lists container images.
func (r *ReplicaSet) ListImages(ctx context.Context, fqn string) ([]string, error) {
	rs, err := r.Load(r.Factory, fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&rs.Spec.Template.Spec), nil
}

// Load returns a given instance.
func (r *ReplicaSet) Load(f Factory, path string) (*appsv1.ReplicaSet, error) {
	o, err := f.Get("apps/v1/replicasets", path, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var rs appsv1.ReplicaSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rs)
	if err != nil {
		return nil, err
	}

	return &rs, nil
}

func getRSRevision(rs *appsv1.ReplicaSet) (int64, error) {
	revision := rs.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
	if rs.Status.Replicas != 0 {
		return 0, errors.New("can not rollback current replica")
	}
	vers, err := strconv.Atoi(revision)
	if err != nil {
		return 0, errors.New("revision conversion failed")
	}

	return int64(vers), nil
}

func controllerInfo(rs *appsv1.ReplicaSet) (string, string, string, error) {
	for _, ref := range rs.ObjectMeta.OwnerReferences {
		if ref.Controller == nil {
			continue
		}
		group, tokens := ref.APIVersion, strings.Split(ref.APIVersion, "/")
		if len(tokens) == 2 {
			group = tokens[0]
		}
		return ref.Name, ref.Kind, group, nil
	}
	return "", "", "", fmt.Errorf("unable to find controller for replicaset: %s", rs.ObjectMeta.Name)
}

// Rollback reverses the last deployment.
func (r *ReplicaSet) Rollback(fqn string) error {
	rs, err := r.Load(r.Factory, fqn)
	if err != nil {
		return err
	}

	version, err := getRSRevision(rs)
	if err != nil {
		return err
	}
	name, kind, apiGroup, err := controllerInfo(rs)
	if err != nil {
		return err
	}
	dial, err := r.Client().Dial()
	if err != nil {
		return err
	}

	rb, err := polymorphichelpers.RollbackerFor(schema.GroupKind{
		Group: apiGroup,
		Kind:  kind,
	},
		dial,
	)
	if err != nil {
		return err
	}

	var ddp Deployment
	ddp.Init(r.Factory, client.NewGVR("apps/v1/deployments"))
	dp, err := ddp.GetInstance(client.FQN(rs.Namespace, name))
	if err != nil {
		return err
	}

	_, err = rb.Rollback(dp, map[string]string{}, version, cmdutil.DryRunNone)
	if err != nil {
		return err
	}

	return nil
}
