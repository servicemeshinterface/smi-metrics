package istio

import (
	"github.com/prometheus/common/log"

	"github.com/deislabs/smi-metrics/pkg/mesh"
	"github.com/deislabs/smi-metrics/pkg/prometheus"
	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/prometheus/common/model"
	v1 "k8s.io/api/core/v1"
)

type edgeLookup struct {
	Item     *metrics.TrafficMetricsList
	interval *metrics.Interval
	details  mesh.ResourceDetails
	queries  map[string]string
}

func (e *edgeLookup) Get(labels model.Metric) *metrics.TrafficMetrics {

	// TODO: test for result labels to have *all* requirements and throw error
	// otherwise (throw in Client.Update)
	var edge *metrics.Edge
	src, dst, err := GetObjectsReference(labels)
	if err != nil {
		log.Error(err)
		return nil
	}
	if src.Name == e.Item.Resource.Name {
		edge = &metrics.Edge{
			Direction: metrics.To,
			Resource: &v1.ObjectReference{
				Kind:      dst.Kind,
				Namespace: dst.Namespace,
				Name:      dst.Name,
			},
		}
	} else if dst.Name == e.Item.Resource.Name {
		edge = &metrics.Edge{
			Direction: metrics.From,
			Resource: &v1.ObjectReference{
				Kind:      src.Kind,
				Namespace: src.Namespace,
				Name:      src.Name,
			},
		}
	}

	obj := e.Item.Get(mesh.ListKey(
		e.Item.Resource.Kind,
		e.Item.Resource.Name,
		e.Item.Resource.Namespace,
	), edge.Resource)
	obj.Interval = e.interval
	obj.Edge = edge

	return obj
}

func (e *edgeLookup) Queries() []*prometheus.Query {
	queries := []*prometheus.Query{}
	for name, tmpl := range e.queries {
		queries = append(queries, &prometheus.Query{
			Name:     name,
			Template: tmpl,
			Values: map[string]interface{}{
				"kind":      e.Item.Resource.Kind,
				"namespace": e.Item.Resource.Namespace,
				"toName":    e.Item.Resource.Name,
			},
		})
	}

	for name, tmpl := range e.queries {
		queries = append(queries, &prometheus.Query{
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
