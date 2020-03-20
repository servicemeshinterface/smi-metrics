package prometheus

import (
	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/mesh"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

type getResourceFunc func(r *ResourceLookup, labels model.Metric) *v1.ObjectReference

type ResourceLookup struct {
	Item        *metrics.TrafficMetricsList
	interval    *metrics.Interval
	queries     map[string]string
	getResource getResourceFunc
}

func newResourceLookup(item *metrics.TrafficMetricsList,
	interval *metrics.Interval,
	queries map[string]string,
	getResource getResourceFunc) *ResourceLookup {

	return &ResourceLookup{
		Item:        item,
		interval:    interval,
		queries:     queries,
		getResource: getResource,
	}
}

func (r *ResourceLookup) Get(labels model.Metric) *metrics.TrafficMetrics {

	result := r.getResource(r, labels)

	// Traffic Metrics Object
	obj := r.Item.Get(mesh.ListKey(
		result.Kind,
		result.Name,
		result.Namespace,
	), nil)
	obj.Interval = r.interval
	obj.Edge = &metrics.Edge{
		Direction: metrics.From,
	}
	return obj
}

func (r *ResourceLookup) Queries() []*Query {
	queries := []*Query{}
	for name, tmpl := range r.queries {
		queries = append(queries, &Query{
			Name:     name,
			Template: tmpl,
			Values: map[string]interface{}{
				"kind":      r.Item.Resource.Kind,
				"namespace": r.Item.Resource.Namespace,
				"name":      r.Item.Resource.Name,
			},
		})
	}

	return queries
}
