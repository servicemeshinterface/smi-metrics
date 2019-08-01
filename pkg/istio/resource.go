package istio

import (
	"strings"

	"github.com/deislabs/smi-metrics/pkg/mesh"

	"github.com/deislabs/smi-metrics/pkg/prometheus"
	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/prometheus/common/model"
)

type resourceLookup struct {
	Item     *metrics.TrafficMetricsList
	interval *metrics.Interval
	queries  map[string]string
}

func (r *resourceLookup) Get(labels model.Metric) *metrics.TrafficMetrics {
	destOwner, ok := labels["destination_owner"]

	if ok {
		// Example Value "kubernetes://apis/apps/v1/namespaces/emojivoto/deployments/voting"
		values := strings.Split(string(destOwner), "/")
		obj := r.Item.Get(mesh.ListKey(
			r.Item.Resource.Kind,
			values[len(values)-1],
			values[len(values)-3],
		), nil)
		obj.Interval = r.interval
		obj.Edge = &metrics.Edge{
			Direction: metrics.From,
		}
		return obj
	} else {

		//Namespace Query
		obj := r.Item.Get(mesh.ListKey(
			r.Item.Resource.Kind,
			r.Item.Resource.Name,
			"",
		), nil)
		obj.Interval = r.interval
		obj.Edge = &metrics.Edge{
			Direction: metrics.From,
		}
		return obj
	}
}

func (r *resourceLookup) Queries() []*prometheus.Query {
	queries := []*prometheus.Query{}
	for name, tmpl := range r.queries {
		queries = append(queries, &prometheus.Query{
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
