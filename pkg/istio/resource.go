package istio

import (
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	v1 "k8s.io/api/core/v1"
)

// Takes an Prometheus result and gives back a kubernetes object reference
func getResource(r *prometheus.ResourceLookup, labels model.Metric) *v1.ObjectReference {

	var result *v1.ObjectReference
	src, dst, err := GetObjectsReference(labels)
	if err != nil {
		log.Error(err)
		return nil
	}
	if src == nil {
		result = dst
	} else {
		result = src
	}

	return result
}
