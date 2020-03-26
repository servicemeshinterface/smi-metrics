package istio

import (
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// Takes an Item i.e the querying Object along with prometheus result and gives back an edge object
func getEdge(e *prometheus.EdgeLookup, labels model.Metric) *metrics.Edge {
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
	return edge
}
