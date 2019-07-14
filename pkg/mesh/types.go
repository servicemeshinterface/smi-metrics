package mesh

import (
	"context"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Mesh interface {
	// Return the list of supported resource kinds
	GetSupportedResources(ctx context.Context) (*metav1.APIResourceList, error)
	// Return metrics for a resource or for the type if name is empty
	GetResourceMetrics(ctx context.Context,
		name, namespace, kind string,
		interval *metrics.Interval) (*metrics.TrafficMetricsList, error)
	// Return the Edge Metrics for a resource
	GetEdgeMetrics(ctx context.Context,
		name, namespace, kind string,
		interval *metrics.Interval,
		details *ResourceDetails) (*metrics.TrafficMetricsList, error)
}

type Queries struct {
	ResourceQueries map[string]string `yaml:"resourceQueries"`
	EdgeQueries     map[string]string `yaml:"edgeQueries"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
