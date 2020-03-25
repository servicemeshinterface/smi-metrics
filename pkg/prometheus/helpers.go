package prometheus

import (
	"context"

	"github.com/deislabs/smi-metrics/pkg/mesh"
	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"
)

type Queries struct {
	ResourceQueries map[string]string `yaml:"resourceQueries"`
	EdgeQueries     map[string]string `yaml:"edgeQueries"`
}

func GetResourceTrafficMetricsList(ctx context.Context,
	obj *v1.ObjectReference,
	interval *metrics.Interval,
	queries map[string]string,
	client promv1.API,
	getResource getResourceFunc,
	getRoute getRouteFunc) (*metrics.TrafficMetricsList, error) {
	// Get is somewhat of a special case as *most* handlers just return a list.
	// Create a list with a fully specified object reference and then just
	// return a single element to keep the code as similar as possible.
	lookup := newResourceLookup(metrics.NewTrafficMetricsList(obj, false),
		interval,
		queries,
		getResource,
		getRoute)

	if err := NewClient(ctx, client, interval).Update(
		lookup); err != nil {
		log.Error(err)
		return nil, err
	}
	return lookup.Item, nil
}

func GetEdgeTraffifMetricsList(ctx context.Context,
	obj *v1.ObjectReference,
	interval *metrics.Interval,
	details *mesh.ResourceDetails,
	queries map[string]string,
	client promv1.API,
	getEdge getEdgeFunc) (*metrics.TrafficMetricsList, error) {

	lookup := newEdgeLookup(metrics.NewTrafficMetricsList(obj, true),
		interval,
		*details,
		queries,
		getEdge)

	if err := NewClient(ctx, client, interval).Update(
		lookup); err != nil {
		return nil, err
	}
	return lookup.Item, nil
}
