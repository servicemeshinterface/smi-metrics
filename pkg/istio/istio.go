package istio

import (
	"context"
	"fmt"

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
	PrometheusURL    string  `yaml:"prometheusUrl"`
	NamespaceQueries Queries `yaml:"namespaceQueries"`
	PodQueries       Queries `yaml:"podQueries"`
	WorkloadQueries  Queries `yaml:"workloadQueries"`
}

type Queries struct {
	ResourceQueries map[string]string `yaml:"resourceQueries"`
	EdgeQueries     map[string]string `yaml:"edgeQueries"`
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

	log.Info(fmt.Sprintf("Query for %s/%s/%s", query.Namespace, query.Kind, query.Name))
	lookup := &edgeLookup{
		Item: metrics.NewTrafficMetricsList(&v1.ObjectReference{
			Kind: query.Kind,
			Name: query.Name,
			// If a namespace isn't defined, it'll be the empty string which fits
			// with the struct's idea of "empty"
			Namespace: query.Namespace,
		}, true),
		details:  *details,
		interval: interval,
		queries:  queries,
	}

	if err := prometheus.NewClient(ctx, l.prometheusClient, interval).Update(
		lookup); err != nil {
		return nil, err
	}
	return lookup.Item, nil
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
	// Get is somewhat of a special case as *most* handlers just return a list.
	// Create a list with a fully specified object reference and then just
	// return a single element to keep the code as similar as possible.
	lookup := &resourceLookup{
		Item:     metrics.NewTrafficMetricsList(obj, false),
		interval: interval,
		queries:  queries,
	}

	if err := prometheus.NewClient(ctx, l.prometheusClient, interval).Update(
		lookup); err != nil {
		log.Error(err)
		return nil, err
	}

	return lookup.Item, nil
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
