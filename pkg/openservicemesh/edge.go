package openservicemesh

import (
	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
)

func getEdge(e *prometheus.EdgeLookup, labels model.Metric) *metrics.Edge {
	src := getSrcObjectRef(e.Item, labels)
	dst := getDestObjectRef(e.Item, labels)

	var edge *metrics.Edge
	if src.Name == e.Item.Resource.Name &&
		src.Namespace == e.Item.Resource.Namespace {
		edge = &metrics.Edge{
			Direction: metrics.To,
			Resource:  dst,
		}
	} else if dst.Name == e.Item.Resource.Name &&
		dst.Namespace == e.Item.Resource.Namespace {
		edge = &metrics.Edge{
			Direction: metrics.From,
			Resource:  src,
		}
	}

	if edge != nil && !e.Details.Namespaced {
		edge.Resource.Namespace = ""
	}

	return edge
}
