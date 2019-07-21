package prometheus

import (
	"strings"

	"github.com/deislabs/smi-metrics/pkg/mesh"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/prometheus/common/model"
)

type ResourceLookup struct {
	Item        *metrics.TrafficMetricsList
	Interval    *metrics.Interval
	PromQueries map[string]string
}

func (r *ResourceLookup) Get(labels model.Metric) *metrics.TrafficMetrics {
	labelName := model.LabelName(strings.ToLower(r.Item.Resource.Kind))

	obj := r.Item.Get(mesh.ListKey(
		r.Item.Resource.Kind,
		string(labels[labelName]),
		string(labels["namespace"]),
	), nil)
	obj.Interval = r.Interval
	obj.Edge = &metrics.Edge{
		Direction: metrics.From,
	}

	return obj
}

func (r *ResourceLookup) Queries() []*Query {
	queries := []*Query{}
	for name, tmpl := range r.PromQueries {
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
