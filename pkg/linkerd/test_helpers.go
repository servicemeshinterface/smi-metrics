package linkerd

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path"
	"time"

	"github.com/prometheus/common/model"
	"github.com/servicemeshinterface/smi-metrics/pkg/linkerd/mocks"
	"github.com/servicemeshinterface/smi-metrics/pkg/metrics"
	"github.com/servicemeshinterface/smi-metrics/pkg/prometheus"
	smimetrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	metricNames = map[string]bool{
		"p99_response_latency": true,
		"p90_response_latency": true,
		"p50_response_latency": true,
		"success_count":        true,
		"failure_count":        true,
	}

	sampleData = []testData{
		{
			name:      "prometheus",
			namespace: "default",
		},
		{
			name:      "foobar",
			namespace: "other",
		},
		{
			name:      "kube-proxy",
			namespace: "kube-system",
		},
	}
)

type kind string

const (
	namespaceKind kind = "Namespace"
)

type testData struct {
	name      string
	namespace string
}

// Suite is a testing suite with a mock client and a handler instance
type Suite struct {
	suite.Suite

	handler *metrics.Handler
	client  *mocks.API

	groupVersion string
}

type apiTest struct {
	Client   *mocks.API
	Suite    *suite.Suite
	Snippets []string
}

func (a *apiTest) MatchQueryParam() func(string) bool {
	assert := a.Suite.Assert()

	return func(query string) bool {
		for _, snippet := range a.Snippets {
			assert.Regexp(snippet, query)
		}

		return true
	}
}

func (a *apiTest) MockQuery(
	kind string,
	labels []model.Metric,
	num int) (float64, *smimetrics.Interval) {
	result := model.Vector{}

	window := 30 * time.Second
	val := rand.Float64()
	interval := &smimetrics.Interval{
		Timestamp: metav1.NewTime(time.Now()),
		Window:    metav1.Duration{Duration: window},
	}

	for _, labelSet := range labels {
		result = append(result, &model.Sample{
			Metric:    labelSet,
			Value:     model.SampleValue(val),
			Timestamp: model.Time(interval.Timestamp.Time.Unix()),
		})
	}

	a.Client.On(
		"Query",
		mock.Anything,
		mock.MatchedBy(a.MatchQueryParam()),
		mock.Anything).Return(result, nil).Times(num)

	return val, interval
}

func (s *Suite) validateTrafficMetrics(
	kind string,
	val float64,
	interval *smimetrics.Interval,
	sample testData,
	result *smimetrics.TrafficMetrics) {

	assert := s.Assert()
	require := s.Require()

	// TypeMeta
	assert.Equal(smimetrics.APIVersion, result.TypeMeta.APIVersion)
	assert.Equal("TrafficMetrics", result.TypeMeta.Kind)

	// ObjectMeta
	r, ok := smimetrics.AvailableKinds[result.Resource.Kind]
	require.Truef(ok, "%s should be a valid kind", result.Resource.Kind)

	if r.Namespaced {
		assert.NotEmpty(result.ObjectMeta.Name)
		assert.Equal(sample.namespace, result.ObjectMeta.Namespace)

		assert.Contains(result.ObjectMeta.SelfLink, path.Join(
			sample.namespace,
			r.Name,
			sample.name,
		))
	} else {
		assert.Equal(sample.namespace, result.ObjectMeta.Name)
		assert.Empty(result.ObjectMeta.Namespace)

		assert.Contains(result.ObjectMeta.SelfLink, path.Join(
			r.Name,
			sample.namespace,
		))
	}

	// Interval
	assert.WithinDuration(
		interval.Timestamp.Time,
		result.Interval.Timestamp.Time,
		1*time.Second)
	assert.Equal(interval.Window, result.Interval.Window)

	// Namespaces work a little differently than the rest of the resources and
	// won't contain the name/namespace fields in a predictable manner, just
	// skip these tests for that Kind.
	if result.Resource.Kind != string(namespaceKind) {
		assert.Equal(sample.namespace, result.Resource.Namespace)
	}
	assert.Equal(kind, result.Resource.Kind)

	// Metrics
	for _, metric := range result.Metrics {
		_, ok := metricNames[metric.Name]
		assert.True(ok, "metric should have expected name")

		assert.Equal(apiresource.NewMilliQuantity(
			int64(val*1000), apiresource.DecimalSI).String(),
			metric.Value.String())
	}
}

func (s *Suite) request(
	method, route string) *httptest.ResponseRecorder {
	require := s.Require()

	req, err := http.NewRequest(method, route, nil)
	require.NoError(err)

	routes := s.handler.Routes()

	rr := httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	require.Equalf(
		rr.Code,
		http.StatusOK,
		"request should be ok (%d): \n%s",
		rr.Code,
		rr.Body.String())

	return rr
}

// SetupTest sets up the tests suite with a handler and a mock client
func (s *Suite) SetupTest() {
	s.groupVersion = "testing.k8s.io/v1beta1"

	file, err := ioutil.ReadFile("test_queries.yaml")
	s.Require().NoError(err)

	var queries prometheus.Queries
	err = yaml.Unmarshal(file, &queries)
	s.Require().NoError(err)

	config := Config{
		"http://stub:9090",
		queries.ResourceQueries,
		queries.EdgeQueries,
	}

	linkerdMesh, err := NewLinkerdProvider(config)
	s.Require().NoError(err)

	handler, err := metrics.NewHandler(linkerdMesh)
	s.Require().NoError(err)

	s.client = &mocks.API{}

	handler.Mesh.(*Linkerd).prometheusClient = s.client
	s.handler = handler
}
