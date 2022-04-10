package dao

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/itchyny/gojq"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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

	dashboardsConfig, err := config.LoadDashboard()
	if err != nil {
		return nil, err
	}

	for gvr, dashboardConfig := range dashboardsConfig.GVRs {
		if dashboardConfig.Active {
			dash.AddForGVR(ctx, ns, &table, gvr, dashboardConfig)
		}
	}

	return table, nil
}

func (dash *Dashboard) AddForGVR(ctx context.Context, ns string, table *[]runtime.Object, gvr string, dashboardConfig config.DashboardGVR) {
	gvrDashboard, err := dash.ForGVR(ctx, ns, client.NewGVR(gvr), dashboardConfig)
	if err == nil {
		*table = append(*table, gvrDashboard)
	}
}

func (dash *Dashboard) ForGVR(ctx context.Context, ns string, gvr client.GVR, dashboardConfig config.DashboardGVR) (render.DashboardRes, error) {
	objects, err := dash.Factory.List(gvr.String(), ns, false, labels.Everything())
	if err != nil {
		return render.DashboardRes{}, err
	}

	modifiedQuery := getQuery(dashboardConfig.Colors.Modified)
	addedQuery := getQuery(dashboardConfig.Colors.Added)
	pendingQuery := getQuery(dashboardConfig.Colors.Pending)
	errorsQuery := getQuery(dashboardConfig.Colors.Error)
	stdQuery := getQuery(dashboardConfig.Colors.Std)
	highlightQuery := getQuery(dashboardConfig.Colors.Highlight)
	killQuery := getQuery(dashboardConfig.Colors.Kill)
	completedQuery := getQuery(dashboardConfig.Colors.Completed)

	columnQueries := make(map[string]*gojq.Query)
	columnCounts := make(map[string]int)
	for column, queryString := range dashboardConfig.Columns {
		columnQueries[column] = getQuery(queryString)
		columnCounts[column] = 0
	}

	columnCounts["TOTAL"] = len(objects)
	columnCounts["MODIFIED"] = 0
	columnCounts["ADDED"] = 0
	columnCounts["PENDING"] = 0
	columnCounts["ERROR"] = 0
	columnCounts["STD"] = 0
	columnCounts["HIGHLIGHT"] = 0
	columnCounts["KILL"] = 0
	columnCounts["COMPLETED"] = 0

	for _, o := range objects {
		obj := o.(*unstructured.Unstructured).Object

		if isMatch(modifiedQuery, obj) {
			columnCounts["MODIFIED"]++
		}
		if isMatch(addedQuery, obj) {
			columnCounts["ADDED"]++
		}
		if isMatch(pendingQuery, obj) {
			columnCounts["PENDING"]++
		}
		if isMatch(errorsQuery, obj) {
			columnCounts["ERROR"]++
		}
		if isMatch(stdQuery, obj) {
			columnCounts["STD"]++
		}
		if isMatch(highlightQuery, obj) {
			columnCounts["HIGHLIGHT"]++
		}
		if isMatch(killQuery, obj) {
			columnCounts["KILL"]++
		}
		if isMatch(completedQuery, obj) {
			columnCounts["COMPLETED"]++
		}
		for column, query := range columnQueries {
			if isMatch(query, obj) {
				columnCounts[column]++
			}
		}
	}

	return render.DashboardRes{
		GVR:    gvr,
		Counts: columnCounts,
	}, nil
}

func getQuery(queryString string) *gojq.Query {
	if len(queryString) == 0 {
		return nil
	}
	query, err := gojq.Parse(queryString)
	if err != nil {
		log.Error().Msgf("%s", err)
		return nil
	}
	return query
}

func isMatch(query *gojq.Query, obj interface{}) bool {
	if query == nil {
		return false
	}
	values := make([]interface{}, 0)
	iter := query.Run(obj)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Error().Msgf("%s", err)
		}
		values = append(values, v)
	}

	if len(values) == 0 {
		return false
	}
	if len(values) > 1 {
		return true
	}

	if res, ok := values[0].(bool); !ok {
		return true
	} else {
		return res
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
