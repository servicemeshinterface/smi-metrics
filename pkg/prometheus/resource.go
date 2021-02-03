package prometheus

import (
	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/mesh"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	v1 "k8s.io/api/core/v1"
)

type getResourceFunc func(r *ResourceLookup, labels model.Metric) *v1.ObjectReference
type getBackendFunc func(r *ResourceLookup, labels model.Metric) *metrics.Backend

type ResourceLookup struct {
	Item        *metrics.TrafficMetricsList
	interval    *metrics.Interval
	queries     map[string]string
	values      map[string]string
	getResource getResourceFunc
	getBackend  getBackendFunc
}

func newResourceLookup(item *metrics.TrafficMetricsList,
	interval *metrics.Interval,
	queries,
	values map[string]string,
	getResource getResourceFunc,
	getBackend getBackendFunc) *ResourceLookup {

	return &ResourceLookup{
		Item:        item,
		interval:    interval,
		queries:     queries,
		values:      values,
		getResource: getResource,
		getBackend:  getBackend,
	}
}

func (r *ResourceLookup) Get(labels model.Metric) *metrics.TrafficMetrics {

	result := r.getResource(r, labels)
	var backend *metrics.Backend
	if r.getBackend != nil {
		backend = r.getBackend(r, labels)
	}

	// Traffic Metrics Object
	obj := getMetrics(r.Item, mesh.ListKey(
		result.Kind,
		result.Name,
		result.Namespace,
	), nil, backend)
	obj.Interval = r.interval
	obj.Edge = &metrics.Edge{
		Direction: metrics.From,
	}
	obj.Backend = backend
	return obj
}

func (r *ResourceLookup) Queries() []*Query {
	queries := []*Query{}
	for name, tmpl := range r.queries {
		values := map[string]interface{}{
			"kind":      r.Item.Resource.Kind,
			"namespace": r.Item.Resource.Namespace,
			"name":      r.Item.Resource.Name,
		}
		for k, v := range r.values {
			values[k] = v
		}
		queries = append(queries, &Query{
			Name:     name,
			Template: tmpl,
			Values:   values,
		})
	}

	return queries
}

// getMetrics will get the item that is associated with the object
// reference or create a default if it doesn't already exist.
func getMetrics(lst *metrics.TrafficMetricsList, obj, edge *v1.ObjectReference,
	backend *metrics.Backend) *metrics.TrafficMetrics {

	for _, item := range lst.Items {
		if objMatch(obj, item.Resource) {
			if edge == nil || (item.Edge != nil &&
				item.Edge.Resource != nil &&
				objMatch(edge, item.Edge.Resource)) {
				if backend == nil || (item.Backend != nil &&
					item.Backend.Name == backend.Name) {
					return item
				}
			}
		}
	}

	t := metrics.NewTrafficMetrics(obj, edge)
	t.Backend = backend
	lst.Items = append(lst.Items, t)

	return t
}

func objMatch(left, right *v1.ObjectReference) bool {
	return left.Kind == right.Kind &&
		left.Namespace == right.Namespace &&
		left.Name == right.Name
}
