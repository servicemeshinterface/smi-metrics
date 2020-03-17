package prometheus

import (
	"github.com/deislabs/smi-metrics/pkg/mesh"
	v1 "k8s.io/api/core/v1"

	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	"github.com/prometheus/common/model"
)

type getResourceFunc func(r *ResourceLookup, labels model.Metric) *v1.ObjectReference
type getRouteFunc func(r *ResourceLookup, labels model.Metric) string

type ResourceLookup struct {
	Item        *metrics.TrafficMetricsList
	interval    *metrics.Interval
	queries     map[string]string
	getResource getResourceFunc
	getRoute getRouteFunc
}

func newResourceLookup(item *metrics.TrafficMetricsList,
	interval *metrics.Interval,
	queries map[string]string,
	getResource getResourceFunc,
	getRoute getRouteFunc) *ResourceLookup {

	return &ResourceLookup{
		Item:        item,
		interval:    interval,
		queries:     queries,
		getResource: getResource,
		getRoute: getRoute,
	}
}

func (r *ResourceLookup) Get(labels model.Metric) *metrics.TrafficMetrics {

	result := r.getResource(r, labels)
	route := ""
	if r.getRoute != nil {
		route = r.getRoute(r, labels)
	}

	// Traffic Metrics Object
	obj := getRouteMetrics(r.Item, mesh.ListKey(
		result.Kind,
		result.Name,
		result.Namespace,
	), nil, route)
	obj.Interval = r.interval
	obj.Edge = &metrics.Edge{
		Direction: metrics.From,
	}
	obj.Route = route
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

// getRouteMetrics will get the item that is associated with the object
// reference or create a default if it doesn't already exist.
func getRouteMetrics(lst *metrics.TrafficMetricsList, obj, edge *v1.ObjectReference,
	route string) *metrics.TrafficMetrics {

	for _, item := range lst.Items {
		if objMatch(obj, item.Resource) {
			if edge == nil || (item.Edge != nil &&
				item.Edge.Resource != nil &&
				objMatch(edge, item.Edge.Resource)) {
				if item.Route == route {
					return item
				}
			}
		}
	}

	t := metrics.NewTrafficMetrics(obj, edge)
	t.Route = route
	lst.Items = append(lst.Items, t)

	return t
}

func objMatch(left, right *v1.ObjectReference) bool {
	return left.Kind == right.Kind &&
		left.Namespace == right.Namespace &&
		left.Name == right.Name
}