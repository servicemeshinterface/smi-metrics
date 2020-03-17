package istio

import (
	"context"

	"github.com/deislabs/smi-metrics/pkg/mesh"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deislabs/smi-metrics/pkg/prometheus"
	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	"github.com/prometheus/client_golang/api"
	v1 "k8s.io/api/core/v1"
)

type Config struct {
	PrometheusURL    string             `yaml:"prometheusUrl"`
	NamespaceQueries prometheus.Queries `yaml:"namespaceQueries"`
	PodQueries       prometheus.Queries `yaml:"podQueries"`
	WorkloadQueries  prometheus.Queries `yaml:"workloadQueries"`
}

type Istio struct {
	config           Config
	prometheusClient promv1.API
}

func (l *Istio) GetSupportedResources(ctx context.Context) (*metav1.APIResourceList, error) {

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

func (l *Istio) GetEdgeMetrics(ctx context.Context,
	query mesh.Query,
	interval *metrics.Interval,
	details *mesh.ResourceDetails) (*metrics.TrafficMetricsList, error) {

	var queries map[string]string

	switch query.Kind {
	case "Namespace":
		queries = l.config.NamespaceQueries.EdgeQueries
	case "Pod":
		queries = l.config.PodQueries.EdgeQueries
	default:
		queries = l.config.WorkloadQueries.EdgeQueries
	}
	obj := &v1.ObjectReference{
		Kind:      query.Kind,
		Name:      query.Name,
		Namespace: query.Namespace,
	}

	metricList, err := prometheus.GetEdgeTraffifMetricsList(ctx,
		obj,
		interval,
		details,
		queries,
		l.prometheusClient,
		getEdge)
	if err != nil {
		return nil, err
	}

	return metricList, nil
}

func (l *Istio) GetResourceMetrics(ctx context.Context,
	query mesh.Query,
	interval *metrics.Interval) (*metrics.TrafficMetricsList, error) {

	obj := &v1.ObjectReference{Kind: query.Kind, Namespace: query.Namespace}
	if query.Name != "" {
		obj.Name = query.Name
	}

	var queries map[string]string
	switch query.Kind {
	case Namespace:
		queries = l.config.NamespaceQueries.ResourceQueries
	case Pod:
		queries = l.config.PodQueries.ResourceQueries
	default:
		queries = l.config.WorkloadQueries.ResourceQueries
	}

	metricList, err := prometheus.GetResourceTrafficMetricsList(ctx,
		obj,
		interval,
		queries,
		l.prometheusClient,
		getResource,
		nil)
	if err != nil {
		return nil, err
	}
	return metricList, err
}

func NewIstioProvider(config Config) (*Istio, error) {

	// Creating a Prometheus Client
	promClient, err := api.NewClient(api.Config{Address: config.PrometheusURL})
	if err != nil {
		return nil, err
	}

	return &Istio{
		config:           config,
		prometheusClient: promv1.NewAPI(promClient),
	}, nil
}
