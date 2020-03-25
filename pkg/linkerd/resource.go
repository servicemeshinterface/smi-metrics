package linkerd

import (
	"strings"

	"github.com/deislabs/smi-metrics/pkg/prometheus"

	v1 "k8s.io/api/core/v1"

	"github.com/prometheus/common/model"
)

// Takes an Prometheus result and gives back a kubernetes object reference
func getResource(r *prometheus.ResourceLookup, labels model.Metric) *v1.ObjectReference {
	labelName := model.LabelName(strings.ToLower(r.Item.Resource.Kind))

	return &v1.ObjectReference{
		Kind:      r.Item.Resource.Kind,
		Namespace: string(labels["namespace"]),
		Name:      string(labels[labelName]),
	}
}

func getRoute(r *prometheus.ResourceLookup, labels model.Metric) string {
	route, ok := labels["rt_route"]
	if !ok {
		route = "[DEFAULT]"
	}
	return string(route)
}
