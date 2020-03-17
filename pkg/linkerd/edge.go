package linkerd

import (
	"fmt"
	"strings"

	"github.com/deislabs/smi-metrics/pkg/prometheus"
	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	"github.com/prometheus/common/model"
	v1 "k8s.io/api/core/v1"
)

func getEdge(e *prometheus.EdgeLookup, labels model.Metric) *metrics.Edge {
	kind := strings.ToLower(e.Item.Resource.Kind)
	src := model.LabelName(kind)
	dst := model.LabelName(fmt.Sprintf("dst_%s", kind))

	var edge *metrics.Edge

	if string(labels[src]) == e.Item.Resource.Name {
		edge = &metrics.Edge{
			Direction: metrics.To,
			Resource: &v1.ObjectReference{
				Kind: e.Item.Resource.Kind,
				Name: string(labels[dst]),
			},
		}

		if e.Details.Namespaced {
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

		if e.Details.Namespaced {
			edge.Resource.Namespace = string(
				labels[model.LabelName("namespace")])
		}
	}

	return edge
}
