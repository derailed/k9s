// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultCRDHeader = model1.Header{
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "GROUP"},
	model1.HeaderColumn{Name: "KIND"},
	model1.HeaderColumn{Name: "VERSIONS"},
	model1.HeaderColumn{Name: "SCOPE"},
	model1.HeaderColumn{Name: "ALIASES", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// CustomResourceDefinition renders a K8s CustomResourceDefinition to screen.
type CustomResourceDefinition struct {
	Base
}

// Header returns a header row.
func (c CustomResourceDefinition) Header(_ string) model1.Header {
	return c.doHeader(defaultCRDHeader)
}

// Render renders a K8s resource to screen.
func (c CustomResourceDefinition) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}

	if err := c.defaultRow(raw, row); err != nil {
		return err
	}
	if c.specs.isEmpty() {
		return nil
	}
	cols, err := c.specs.realize(raw, defaultCRDHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

// Render renders a K8s resource to screen.
func (c CustomResourceDefinition) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var crd v1.CustomResourceDefinition
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &crd)
	if err != nil {
		return err
	}

	versions := make([]string, 0, len(crd.Spec.Versions))
	for _, v := range crd.Spec.Versions {
		if v.Served {
			n := v.Name
			if v.Deprecated {
				n += "!"
			}
			versions = append(versions, n)
		}
	}
	if len(versions) == 0 {
		slog.Warn("Unable to assert CRD versions", slogs.FQN, crd.Name)
	}

	r.ID = client.MetaFQN(&crd.ObjectMeta)
	r.Fields = model1.Fields{
		crd.Spec.Names.Plural,
		crd.Spec.Group,
		crd.Spec.Names.Kind,
		naStrings(versions),
		string(crd.Spec.Scope),
		naStrings(crd.Spec.Names.ShortNames),
		mapToIfc(crd.GetLabels()),
		AsStatus(c.diagnose(crd.Name, crd.Spec.Versions)),
		ToAge(crd.GetCreationTimestamp()),
	}

	return nil
}

func (CustomResourceDefinition) diagnose(n string, vv []v1.CustomResourceDefinitionVersion) error {
	if len(vv) == 0 {
		return fmt.Errorf("unable to assert CRD servers versions for %s", n)
	}

	var (
		ee     []error
		served bool
	)
	for _, v := range vv {
		if v.Served {
			served = true
		}
		if v.Deprecated {
			if v.DeprecationWarning != nil {
				ee = append(ee, fmt.Errorf("%s", *v.DeprecationWarning))
			} else {
				ee = append(ee, fmt.Errorf("%s[%s] is deprecated", n, v.Name))
			}
		}
	}
	if !served {
		ee = append(ee, fmt.Errorf("CRD %s is no longer served by the api server", n))
	}

	if len(ee) == 0 {
		return nil
	}
	errs := make([]string, 0, len(ee))
	for _, e := range ee {
		errs = append(errs, e.Error())
	}

	return errors.New(strings.Join(errs, " - "))
}
