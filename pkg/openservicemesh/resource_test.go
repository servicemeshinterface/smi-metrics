package openservicemesh

import (
	"reflect"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func TestGetResource(t *testing.T) {
	tests := []struct {
		name     string
		lookup   *prometheus.ResourceLookup
		labels   model.Metric
		expected *v1.ObjectReference
	}{
		{
			name: "workload",
			lookup: &prometheus.ResourceLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind: "Deployment",
					},
				},
			},
			labels: model.Metric{
				"destination_namespace": "destns",
				"destination_name":      "destname",
			},
			expected: &v1.ObjectReference{
				Kind:      "Deployment",
				Namespace: "destns",
				Name:      "destname",
			},
		},
		{
			name: "pod",
			lookup: &prometheus.ResourceLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind: "Pod",
					},
				},
			},
			labels: model.Metric{
				"destination_namespace": "destns",
				"destination_pod":       "destpod",
			},
			expected: &v1.ObjectReference{
				Kind:      "Pod",
				Namespace: "destns",
				Name:      "destpod",
			},
		},
		{
			name: "namespace",
			lookup: &prometheus.ResourceLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind: "Namespace",
					},
				},
			},
			labels: model.Metric{
				"destination_namespace": "destns",
			},
			expected: &v1.ObjectReference{
				Kind: "Namespace",
				Name: "destns",
			},
		},
		{
			name: "workload with underscore",
			lookup: &prometheus.ResourceLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind: "Deployment",
					},
				},
			},
			labels: model.Metric{
				"destination_namespace": "dest_ns",
				"destination_name":      "dest_name",
			},
			expected: &v1.ObjectReference{
				Kind:      "Deployment",
				Namespace: "dest-ns",
				Name:      "dest-name",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := getResource(test.lookup, test.labels)
			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}
