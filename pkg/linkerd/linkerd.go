package linkerd

import (
	"context"

	"github.com/deislabs/smi-metrics/pkg/mesh"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deislabs/smi-metrics/pkg/prometheus"
	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/prometheus/client_golang/api"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Config struct {
	PrometheusURL   string            `yaml:"prometheusUrl"`
	ResourceQueries map[string]string `yaml:"resourceQueries"`
	EdgeQueries     map[string]string `yaml:"edgeQueries"`
}

type Queries struct {
	ResourceQueries map[string]string `yaml:"resourceQueries"`
	EdgeQueries     map[string]string `yaml:"edgeQueries"`
}

type Linkerd struct {
	queries          Queries
	prometheusClient promv1.API
}

func (l *Linkerd) GetSupportedResources(ctx context.Context) (*metav1.APIResourceList, error) {
	lst := &metav1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResourceList",
			APIVersion: "v1",
		},
		GroupVersion: metrics.APIVersion,
		APIResources: []metav1.APIResource{},
	}

	for _, v := range metrics.AvailableKinds {
		lst.APIResources = append(lst.APIResources, *v)
	}

	return lst, nil
}

func (l *Linkerd) GetEdgeMetrics(ctx context.Context,
	query mesh.Query,
	interval *metrics.Interval,
	details *mesh.ResourceDetails) (*metrics.TrafficMetricsList, error) {

	lookup := &prometheus.EdgeLookup{
		Item: metrics.NewTrafficMetricsList(&v1.ObjectReference{
			Kind: query.Kind,
			Name: query.Name,
			// If a namespace isn't defined, it'll be the empty string which fits
			// with the struct's idea of "empty"
			Namespace: query.Namespace,
		}, true),
		Details:     *details,
		Interval:    interval,
		PromQueries: l.queries.EdgeQueries,
	}

	if err := prometheus.NewClient(ctx, l.prometheusClient, interval).Update(
		lookup); err != nil {
		return nil, err
	}
	return lookup.Item, nil
}

func (l *Linkerd) GetResourceMetrics(ctx context.Context,
	query mesh.Query,
	interval *metrics.Interval) (*metrics.TrafficMetricsList, error) {

	obj := &v1.ObjectReference{Kind: query.Kind, Namespace: query.Namespace}
	if query.Name != "" {
		obj.Name = query.Name
	}

	// Get is somewhat of a special case as *most* handlers just return a list.
	// Create a list with a fully specified object reference and then just
	// return a single element to keep the code as similar as possible.
	lookup := &prometheus.ResourceLookup{
		Item:        metrics.NewTrafficMetricsList(obj, false),
		Interval:    interval,
		PromQueries: l.queries.ResourceQueries,
	}

	if err := prometheus.NewClient(ctx, l.prometheusClient, interval).Update(
		lookup); err != nil {
		log.Error(err)
		return nil, err
	}

	return lookup.Item, nil
}

func NewLinkerdProvider(config Config) (*Linkerd, error) {

	// Creating a Prometheus Client
	promClient, err := api.NewClient(api.Config{Address: config.PrometheusURL})
	if err != nil {
		return nil, err
	}

	queries := Queries{
		ResourceQueries: config.ResourceQueries,
		EdgeQueries:     config.EdgeQueries,
	}

	return &Linkerd{
		queries:          queries,
		prometheusClient: promv1.NewAPI(promClient),
	}, nil
}
