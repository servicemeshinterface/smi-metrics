package linkerd

import (
	"strings"

	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	v1 "k8s.io/api/core/v1"
)

// Takes an Prometheus result and gives back a kubernetes object reference
func getResource(r *prometheus.ResourceLookup, labels model.Metric) *v1.ObjectReference {
	if r.Item.Resource.Kind == trafficsplitKind {
		return r.Item.Resource
	}

	labelName := model.LabelName(strings.ToLower(r.Item.Resource.Kind))

	return &v1.ObjectReference{
		Kind:      r.Item.Resource.Kind,
		Namespace: string(labels["namespace"]),
		Name:      string(labels[labelName]),
	}
}
