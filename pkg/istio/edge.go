package istio

import (
	"fmt"
	"strings"

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
	kind := strings.ToLower(e.Item.Resource.Kind)
	src := model.LabelName(kind)
	dst := model.LabelName(fmt.Sprintf("dst_%s", kind))

	// TODO: test for result labels to have *all* requirements and throw error
	// otherwise (throw in Client.Update)
	var edge *metrics.Edge

	if string(labels[src]) == e.Item.Resource.Name {
		edge = &metrics.Edge{
			Direction: metrics.To,
			Resource: &v1.ObjectReference{
				Kind: e.Item.Resource.Kind,
				Name: string(labels[dst]),
			},
		}

		if e.details.Namespaced {
			edge.Resource.Namespace = string(
				labels[model.LabelName("dst_namespace")])
		}
	} else {
		edge = &metrics.Edge{
			Direction: metrics.From,
			Resource: &v1.ObjectReference{
				Kind: e.Item.Resource.Kind,
				Name: string(labels[src]),
			},
		}

		if e.details.Namespaced {
			edge.Resource.Namespace = string(
				labels[model.LabelName("namespace")])
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