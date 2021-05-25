package openservicemesh

import (
	"reflect"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/mesh"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func TestGetEdge(t *testing.T) {
	tests := []struct {
		name     string
		lookup   *prometheus.EdgeLookup
		labels   model.Metric
		expected *metrics.Edge
	}{
		{
			name: "workload as src",
			lookup: &prometheus.EdgeLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind:      "Deployment",
						Namespace: "srcns",
						Name:      "srcname",
					},
				},
				Details: mesh.ResourceDetails{
					Namespaced: true,
				},
			},
			labels: model.Metric{
				"destination_namespace": "destns",
				"destination_name":      "destname",
				"source_namespace":      "srcns",
				"source_name":           "srcname",
			},
			expected: &metrics.Edge{
				Direction: metrics.To,
				Resource: &v1.ObjectReference{
					Kind:      "Deployment",
					Namespace: "destns",
					Name:      "destname",
				},
			},
		},
		{
			name: "workload as dest",
			lookup: &prometheus.EdgeLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind:      "Deployment",
						Namespace: "destns",
						Name:      "destname",
					},
				},
				Details: mesh.ResourceDetails{
					Namespaced: true,
				},
			},
			labels: model.Metric{
				"destination_namespace": "destns",
				"destination_name":      "destname",
				"source_namespace":      "srcns",
				"source_name":           "srcname",
			},
			expected: &metrics.Edge{
				Direction: metrics.From,
				Resource: &v1.ObjectReference{
					Kind:      "Deployment",
					Namespace: "srcns",
					Name:      "srcname",
				},
			},
		},
		{
			name: "namespace as dest",
			lookup: &prometheus.EdgeLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind: "Namespace",
						Name: "destns",
					},
				},
			},
			labels: model.Metric{
				"destination_namespace": "destns",
				"source_namespace":      "srcns",
			},
			expected: &metrics.Edge{
				Direction: metrics.From,
				Resource: &v1.ObjectReference{
					Kind: "Namespace",
					Name: "srcns",
				},
			},
		},
		{
			name: "pod as dest",
			lookup: &prometheus.EdgeLookup{
				Item: &v1alpha1.TrafficMetricsList{
					Resource: &v1.ObjectReference{
						Kind:      "Pod",
						Namespace: "destns",
						Name:      "destpod",
					},
				},
				Details: mesh.ResourceDetails{
					Namespaced: true,
				},
			},
			labels: model.Metric{
				"destination_namespace": "destns",
				"destination_pod":       "destpod",
				"source_namespace":      "srcns",
				"source_pod":            "srcpod",
			},
			expected: &metrics.Edge{
				Direction: metrics.From,
				Resource: &v1.ObjectReference{
					Kind:      "Pod",
					Namespace: "srcns",
					Name:      "srcpod",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := getEdge(test.lookup, test.labels)
			if test.expected.Direction != actual.Direction {
				t.Errorf("expected direction %s, got %s", test.expected.Direction, actual.Direction)
			}
			if !reflect.DeepEqual(test.expected.Resource, actual.Resource) {
				t.Errorf("expected resource %v, got %v", test.expected.Resource, actual.Resource)
			}
		})
	}
}
