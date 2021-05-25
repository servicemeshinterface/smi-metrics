package openservicemesh

import (
	"strings"

	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// Takes an Prometheus result and gives back a kubernetes object reference.
func getResource(r *prometheus.ResourceLookup, labels model.Metric) *v1.ObjectReference {
	return getDestObjectRef(r.Item, labels)
}

func getDestObjectRef(r *metrics.TrafficMetricsList, labels model.Metric) *v1.ObjectReference {
	return getObjectRef(r, labels, "destination")
}

func getSrcObjectRef(r *metrics.TrafficMetricsList, labels model.Metric) *v1.ObjectReference {
	return getObjectRef(r, labels, "source")
}

func getObjectRef(r *metrics.TrafficMetricsList, labels model.Metric, labelPrefix string) *v1.ObjectReference {
	obj := &v1.ObjectReference{
		Kind:      r.Resource.Kind,
		Namespace: string(labels[model.LabelName(labelPrefix+"_namespace")]),
		Name:      string(labels[model.LabelName(labelPrefix+"_name")]),
	}

	switch r.Resource.Kind {
	case "Namespace":
		obj.Name = string(labels[model.LabelName(labelPrefix+"_namespace")])
		obj.Namespace = ""
	case "Pod":
		obj.Name = string(labels[model.LabelName(labelPrefix+"_pod")])
	}

	// Note: Since OSM uses proxy-wasm to add tags to metrics and those tags
	// show up in the metric name, Envoy converts any '-' or '.' to '_' in
	// metric names for the Prometheus format. This means by looking at the
	// labels on metrics, it's impossible to tell whether a '_' was originally
	// a '.' or '-'. Since '-' seems to be more common, that is chosen in all
	// cases. To work around this issue, queries for resources whose names
	// include a '.' should replace the '.' with '-' in the API query.
	// Namespaces are not affected by this as '.' is not allowed in namespace
	// names, so '_' unambiguously identifies a '-'.
	obj.Name = strings.ReplaceAll(obj.Name, "_", "-")
	obj.Namespace = strings.ReplaceAll(obj.Namespace, "_", "-")

	return obj
}
