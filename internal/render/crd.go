package render

import (
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// CustomResourceDefinition renders a K8s CustomResourceDefinition to screen.
type CustomResourceDefinition struct {
	Base
}

// Header returns a header rbw.
func (CustomResourceDefinition) Header(string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "VERSIONS"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (c CustomResourceDefinition) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected CustomResourceDefinition, but got %T", o)
	}

	var crd v1.CustomResourceDefinition
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &crd)
	if err != nil {
		return err
	}

	versions := make([]string, 0, 3)
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
		log.Warn().Msgf("unable to assert CRD versions for %s", crd.GetName())
	}

	r.ID = client.FQN(client.ClusterScope, crd.GetName())
	r.Fields = Fields{
		crd.GetName(),
		naStrings(versions),
		mapToIfc(crd.GetLabels()),
		asStatus(c.diagnose(crd.GetName(), crd.Spec.Versions)),
		toAge(crd.GetCreationTimestamp()),
	}

	return nil
}

func (c CustomResourceDefinition) diagnose(n string, vv []v1.CustomResourceDefinitionVersion) error {
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
				ee = append(ee, fmt.Errorf("%s[%s] is deprecated!", n, v.Name))
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

func extractMetaField(m map[string]interface{}, field string) string {
	f, ok := m[field]
	if !ok {
		log.Error().Err(fmt.Errorf("failed to extract field from meta %s", field))
		return NAValue
	}

	fs, ok := f.(string)
	if !ok {
		log.Error().Err(fmt.Errorf("failed to extract string from field %s", field))
		return NAValue
	}

	return fs
}
