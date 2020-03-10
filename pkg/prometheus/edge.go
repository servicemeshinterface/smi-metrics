package prometheus

import (
	"github.com/deislabs/smi-metrics/pkg/mesh"
	"github.com/prometheus/common/model"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
)

type getEdgeFunc func(e *EdgeLookup, labels model.Metric) *metrics.Edge

type EdgeLookup struct {
	Item     *metrics.TrafficMetricsList
	interval *metrics.Interval
	Details  mesh.ResourceDetails
	queries  map[string]string
	getEdge  getEdgeFunc
}

func newEdgeLookup(item *metrics.TrafficMetricsList,
	interval *metrics.Interval,
	details mesh.ResourceDetails,
	queries map[string]string,
	edgeFunc getEdgeFunc) *EdgeLookup {

	return &EdgeLookup{
		Item:     item,
		interval: interval,
		Details:  details,
		queries:  queries,
		getEdge:  edgeFunc,
	}
}

func (e *EdgeLookup) Get(labels model.Metric) *metrics.TrafficMetrics {

	edge := e.getEdge(e, labels)

	obj := e.Item.Get(mesh.ListKey(
		e.Item.Resource.Kind,
		e.Item.Resource.Name,
		e.Item.Resource.Namespace,
	), edge.Resource)
	obj.Interval = e.interval
	obj.Edge = edge

	return obj
}

func (e *EdgeLookup) Queries() []*Query {
	queries := []*Query{}
	for name, tmpl := range e.queries {
		queries = append(queries, &Query{
			Name:     name,
			Template: tmpl,
			Values: map[string]interface{}{
				"kind":      e.Item.Resource.Kind,
				"toNamespace": e.Item.Resource.Namespace,
				"toName":    e.Item.Resource.Name,
			},
		})
	}

	for name, tmpl := range e.queries {
		queries = append(queries, &Query{
			Name:     name,
			Template: tmpl,
			Values: map[string]interface{}{
				"kind":      e.Item.Resource.Kind,
				"namespace": e.Item.Resource.Namespace,
				"fromName":  e.Item.Resource.Name,
			},
		})
	}

	return queries
}
