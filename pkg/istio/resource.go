package istio

import (
	"github.com/deislabs/smi-metrics/pkg/prometheus"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"

	"github.com/prometheus/common/model"
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
