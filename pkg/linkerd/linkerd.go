package linkerd

import (
	"context"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/servicemeshinterface/smi-metrics/pkg/mesh"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	PrometheusURL   string            `yaml:"prometheusUrl"`
	ResourceQueries map[string]string `yaml:"resourceQueries"`
	EdgeQueries     map[string]string `yaml:"edgeQueries"`
}

type Linkerd struct {
	queries          prometheus.Queries
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

	obj := &v1.ObjectReference{
		Kind:      query.Kind,
		Name:      query.Name,
		Namespace: query.Namespace,
	}

	metricList, err := prometheus.GetEdgeTraffifMetricsList(ctx,
		obj,
		interval,
		details,
		l.queries.EdgeQueries,
		l.prometheusClient, getEdge)
	if err != nil {
		return nil, err
	}

	return metricList, nil
}

func (l *Linkerd) GetResourceMetrics(ctx context.Context,
	query mesh.Query,
	interval *metrics.Interval) (*metrics.TrafficMetricsList, error) {

	obj := &v1.ObjectReference{Kind: query.Kind, Namespace: query.Namespace}
	if query.Name != "" {
		obj.Name = query.Name
	}

	metricsList, err := prometheus.GetResourceTrafficMetricsList(ctx,
		obj,
		interval,
		l.queries.ResourceQueries,
		l.prometheusClient,
		getResource)
	if err != nil {
		return nil, err
	}
	return metricsList, nil
}

func NewLinkerdProvider(config Config) (*Linkerd, error) {

	// Creating a Prometheus Client
	promClient, err := api.NewClient(api.Config{Address: config.PrometheusURL})
	if err != nil {
		return nil, err
	}

	queries := prometheus.Queries{
		ResourceQueries: config.ResourceQueries,
		EdgeQueries:     config.EdgeQueries,
	}

	return &Linkerd{
		queries:          queries,
		prometheusClient: promv1.NewAPI(promClient),
	}, nil
}
