package linkerd

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/deislabs/smi-metrics/pkg/mesh"
	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
)

type ResourceTestSuite struct {
	Suite
}

func (s *ResourceTestSuite) validateResource(
	sample testData,
	result *metrics.TrafficMetrics) {
	assert := s.Assert()
	require := s.Require()

	// ObjectMeta
	r, ok := metrics.AvailableKinds[result.Resource.Kind]
	require.Truef(ok, "%s should be a valid kind", result.Resource.Kind)

	if r.Namespaced {
		assert.Equal(sample.name, result.ObjectMeta.Name)
	}

	if result.Resource.Kind != string(namespaceKind) {
		assert.Equal(sample.name, result.Resource.Name)
		assert.Equal(sample.namespace, result.Resource.Namespace)
	}

	assert.Equal(metrics.From, result.Edge.Direction)
	assert.Nil(result.Edge.Resource)
}

func (s *ResourceTestSuite) TestListByKind() {
	for kind := range metrics.AvailableKinds {
		kind := kind
		kindLabel := model.LabelName(strings.ToLower(kind))

		tester := &apiTest{
			Client: s.client,
			Suite:  &s.Suite.Suite,
			Snippets: []string{
				fmt.Sprintf("%s=~", strings.ToLower(kind)),
				"namespace=~\".+\"",
				"[30s]",
				fmt.Sprintf(`by\s+\(\s+%s,\s+namespace`, strings.ToLower(kind)),
			},
		}

		var samples []model.Metric
		for _, data := range sampleData {
			samples = append(samples, model.Metric{
				kindLabel:   model.LabelValue(data.name),
				"namespace": model.LabelValue(data.namespace),
			})
		}

		val, interval := tester.MockQuery(kind, samples, 5)

		s.Run(kind, func() {
			assert := s.Assert()
			require := s.Require()

			pth, ok := metrics.AvailableKinds[kind]
			require.Truef(ok, "%s should be available", kind)

			rr := s.request("GET", path.Join("/", pth.Name))

			var resp *metrics.TrafficMetricsList

			assert.NoError(json.Unmarshal(rr.Body.Bytes(), &resp))

			assert.Equal("TrafficMetricsList", resp.TypeMeta.Kind)
			assert.Equal(metrics.APIVersion, resp.TypeMeta.APIVersion)

			assert.Contains(resp.ListMeta.SelfLink, pth.Name)

			assert.Equal(kind, resp.Resource.Kind)
			assert.Empty(resp.Resource.Namespace)
			assert.Empty(resp.Resource.Name)

			assert.Len(resp.Items, 3)

			for i, sample := range sampleData {
				sample := sample

				s.validateTrafficMetrics(
					kind, val, interval, sample, resp.Items[i])
				s.validateResource(sample, resp.Items[i])
			}
		})
	}
}

func (s *ResourceTestSuite) TestGetNamespace() {
	assert := s.Assert()

	details, ok := mesh.GetResourceDetails("namespaces")
	assert.True(ok, "namespaces need to be supported")

	sample := testData{
		name:      "default",
		namespace: "default",
	}

	tester := &apiTest{
		Client:   s.client,
		Suite:    &s.Suite.Suite,
		Snippets: []string{},
	}

	val, interval := tester.MockQuery(details.Kind, []model.Metric{
		{
			"namespace": model.LabelValue(sample.name),
		},
	}, 5)

	rr := s.request("GET", path.Join(
		"/", "namespaces", sample.namespace))

	var resp *metrics.TrafficMetrics

	assert.NoError(json.Unmarshal(rr.Body.Bytes(), &resp))

	assert.Equal("TrafficMetrics", resp.TypeMeta.Kind)
	assert.Equal(metrics.APIVersion, resp.TypeMeta.APIVersion)

	assert.Contains(resp.ObjectMeta.SelfLink, sample.namespace)

	assert.IsType(&v1.ObjectReference{}, resp.Resource)
	assert.Equal(details.Kind, resp.Resource.Kind)
	assert.Empty(resp.Resource.Namespace)
	assert.Equal(sample.namespace, resp.Resource.Name)

	s.validateTrafficMetrics(details.Kind, val, interval, sample, resp)
	s.validateResource(sample, resp)
}

func (s *ResourceTestSuite) TestListByNamespaceKind() {

}

func (s *ResourceTestSuite) TestGet() {
	for _, sample := range sampleData {
		sample := sample

		for kind := range metrics.AvailableKinds {
			kind := kind
			lowerKind := strings.ToLower(kind)
			kindLabel := model.LabelName(strings.ToLower(kind))

			s.Run(kind, func() {
				tester := &apiTest{
					Client: s.client,
					Suite:  &s.Suite.Suite,
					Snippets: []string{
						fmt.Sprintf("%s=~\"%s\"", lowerKind, sample.name),
						fmt.Sprintf("namespace=~\"%s\"", sample.namespace),
						"[30s]",
						fmt.Sprintf(`by\s+\(\s+%s,(\s+rt_route,)?\s+namespace`, lowerKind),
					},
				}

				val, interval := tester.MockQuery(kind, []model.Metric{
					{
						kindLabel:   model.LabelValue(sample.name),
						"namespace": model.LabelValue(sample.namespace),
					},
				}, 5)

				assert := s.Assert()
				require := s.Require()

				pth, ok := metrics.AvailableKinds[kind]
				require.Truef(ok, "%s should be available", kind)

				rr := s.request("GET", path.Join(
					"/",
					"namespaces",
					sample.namespace,
					pth.Name,
					sample.name))

				var resp *metrics.TrafficMetrics

				assert.NoError(json.Unmarshal(rr.Body.Bytes(), &resp))

				assert.Equal("TrafficMetrics", resp.TypeMeta.Kind)
				assert.Equal(metrics.APIVersion, resp.TypeMeta.APIVersion)

				assert.Equal(kind, resp.Resource.Kind)

				s.validateTrafficMetrics(kind, val, interval, sample, resp)
				s.validateResource(sample, resp)
			})
		}
	}
}

func TestResource(t *testing.T) {
	suite.Run(t, new(ResourceTestSuite))
}
