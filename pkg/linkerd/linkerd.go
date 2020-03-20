package linkerd

import (
	"context"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/cluster"
	"github.com/servicemeshinterface/smi-metrics/pkg/mesh"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	trafficsplit "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	tsClient "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	PrometheusURL       string            `yaml:"prometheusUrl"`
	ResourceQueries     map[string]string `yaml:"resourceQueries"`
	EdgeQueries         map[string]string `yaml:"edgeQueries"`
	TrafficSplitQueries map[string]string `yaml:"trafficSplitQueries"`
}

type Linkerd struct {
	queries          Config
	prometheusClient promv1.API
	tsAPI            tsClient.Interface
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
		nil,
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

	queries := l.queries.ResourceQueries
	values := map[string]string{}

	var getBackend func(r *prometheus.ResourceLookup, labels model.Metric) *metrics.Backend

	if query.Kind == "Trafficsplit" {

		ts, err := l.getTrafficSplit(query)
		if err != nil {
			return nil, err
		}

		getBackend = mkGetBackend(ts)
		queries = l.queries.TrafficSplitQueries
		values["apex"] = ts.Spec.Service
	}

	metricsList, err := prometheus.GetResourceTrafficMetricsList(ctx,
		obj,
		interval,
		queries,
		values,
		l.prometheusClient,
		getResource,
		getBackend)
	if err != nil {
		return nil, err
	}
	return metricsList, nil
}

func (l *Linkerd) getTrafficSplit(query mesh.Query) (*trafficsplit.TrafficSplit, error) {
	return l.tsAPI.SplitV1alpha1().TrafficSplits(query.Namespace).Get(query.Name, metav1.GetOptions{})
}

func mkGetBackend(ts *trafficsplit.TrafficSplit) func(r *prometheus.ResourceLookup, labels model.Metric) *metrics.Backend {
	return func(r *prometheus.ResourceLookup, labels model.Metric) *metrics.Backend {
		for _, backend := range ts.Spec.Backends {
			if backend.Service == string(labels["dst_service"]) {
				return &metrics.Backend{
					Apex:   ts.Spec.Service,
					Name:   backend.Service,
					Weight: int(backend.Weight.MilliValue()),
				}
			}
		}
		return nil
	}
}

func NewLinkerdProvider(config Config) (*Linkerd, error) {

	// Creating a Prometheus Client
	promClient, err := api.NewClient(api.Config{Address: config.PrometheusURL})
	if err != nil {
		return nil, err
	}

	kubeconfig, err := cluster.GetKubeconfig()
	if err != nil {
		return nil, err
	}

	tsAPI, err := tsClient.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &Linkerd{
		queries:          config,
		prometheusClient: promv1.NewAPI(promClient),
		tsAPI:            tsAPI,
	}, nil
}
