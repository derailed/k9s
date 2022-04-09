package dao

import (
	"context"
	"reflect"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Dashboard)(nil)

// Dashboard tracks cluster resource status.
type Dashboard struct {
	NonResource
}

// NewDashboard returns a new set of Dashboard items.
func NewDashboard(f Factory) *Dashboard {
	a := Dashboard{}
	a.Init(f, client.NewGVR("dashboard"))

	return &a
}

// List returns a collection of Dashboard items.
func (dash *Dashboard) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	var table []runtime.Object

	dashboardConfig, err := config.LoadDashboard()
	if err != nil {
		return nil, err
	}

	for _, gvr := range dashboardConfig.GVRs {
		dash.AddForGVR(ctx, ns, &table, gvr)
	}

	return table, nil
}

func (dash *Dashboard) AddForGVR(ctx context.Context, ns string, table *[]runtime.Object, gvr string) {
	deployments, err := dash.ForGVR(ctx, ns, client.NewGVR(gvr))
	if err == nil {
		*table = append(*table, deployments)
	}
}

func (dash *Dashboard) ForGVR(ctx context.Context, ns string, gvr client.GVR) (render.DashboardRes, error) {
	IsHappy := dash.GetHasHappyAccessor(gvr)
	objects, err := dash.Factory.List(gvr.String(), ns, false, labels.Everything())
	if err != nil {
		return render.DashboardRes{}, err
	}

	total := len(objects)
	ok := 0
	errors := 0

	for _, o := range objects {
		var obj ApiObject
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &obj)
		if err != nil {
			continue
		}

		var status bool

		if IsHappy != nil {
			status = IsHappy(o)
		} else {
			status = true
			for _, c := range obj.Status.Conditions {
				conditionStatus, err := strconv.ParseBool(string(c.Status))
				if err != nil {
					conditionStatus = false
				}
				status = status && conditionStatus
			}
		}

		if status {
			ok++
		} else {
			errors++
		}
	}
	return render.DashboardRes{GVR: gvr, Total: total, OK: ok, Errors: errors}, nil
}

func (dash *Dashboard) GetHasHappyAccessor(gvr client.GVR) func(runtime.Object) bool {
	accessor, err := AccessorFor(dash.Factory, gvr)
	if err != nil {
		return nil
	}

	isHappy := reflect.ValueOf(accessor).MethodByName("IsHappy")
	if !reflect.Value.IsValid(isHappy) {
		return nil
	}

	argType := isHappy.Type().In(0)

	return func(o runtime.Object) bool {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Msgf("IsHappy error %s", r)
			}
		}()

		obj := reflect.New(argType).Elem()
		err = FromUnstructured(reflect.ValueOf(o.(*unstructured.Unstructured).Object), obj)
		if err != nil {
			log.Error().Msgf("Conversion error %s", err)
			return false
		}

		args := []reflect.Value{obj}
		res := isHappy.Call(args)
		return res[0].Bool()
	}
}

type ApiObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Status            ApiObjectStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type ApiObjectStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}
