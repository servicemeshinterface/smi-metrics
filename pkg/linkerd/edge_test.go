package linkerd

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

var (
	edgeRoot = testData{
		name:      "prometheus",
		namespace: "default",
	}
	edgeData = []edgeTestData{
		{
			direction: metrics.To,
			resource: testData{
				name:      "web",
				namespace: "baz",
			},
		},
		{
			direction: metrics.To,
			resource: testData{
				name:      "foobar",
				namespace: "other",
			},
		},
		{
			direction: metrics.From,
			resource: testData{
				name:      "kube-proxy",
				namespace: "kube-system",
			},
		},
	}
)

type edgeTestData struct {
	direction metrics.Direction
	resource  testData
}

type EdgeTestSuite struct {
	Suite
}

func (s *EdgeTestSuite) validateEdge(
	kind string,
	data edgeTestData,
	resp *metrics.TrafficMetrics,
	selfLink string) {
	assert := s.Assert()
	require := s.Require()

	pth, ok := metrics.AvailableKinds[kind]
	require.Truef(ok, "%s should be available", kind)

	assert.Contains(resp.ObjectMeta.SelfLink, selfLink)

	assert.Equal(data.direction, resp.Edge.Direction)

	if pth.Namespaced {
		assert.Equal(edgeRoot.name, resp.Resource.Name)
		assert.Equal(edgeRoot.namespace, resp.Resource.Namespace)

		assert.Equal(data.resource.name, resp.Edge.Resource.Name)
		assert.Equalf(
			data.resource.namespace,
			resp.Edge.Resource.Namespace,
			resp.Edge.Resource.String())
	} else {
		assert.Equal(edgeRoot.namespace, resp.Resource.Name)
		assert.Empty(resp.Resource.Namespace)

		assert.Equal(data.resource.namespace, resp.Edge.Resource.Name)
		assert.Empty(resp.Edge.Resource.Namespace)
	}
}

func (s *EdgeTestSuite) testCase(kind string) {
	assert := s.Assert()
	require := s.Require()

	lowerKind := strings.ToLower(kind)
	kindLabel := model.LabelName(lowerKind)
	dstKindLabel := model.LabelName(
		fmt.Sprintf("dst_%s", lowerKind))

	tester := &apiTest{
		Client: s.client,
		Suite:  &s.Suite.Suite,
		Snippets: []string{
			fmt.Sprintf(`%s=~`, lowerKind),
			"[30s]",
			fmt.Sprintf(
				`by\s+\(\s+%s,\s+dst_%s,\s+namespace,\s+dst_namespace`,
				lowerKind,
				lowerKind),
		},
	}

	pth, ok := metrics.AvailableKinds[kind]
	require.Truef(ok, "%s should be available", kind)

	var samples []model.Metric
	for _, data := range edgeData {
		var src, dst testData

		if data.direction == metrics.From {
			src = data.resource
			dst = edgeRoot
		} else {
			src = edgeRoot
			dst = data.resource
		}

		var name string
		var dstName string
		if pth.Namespaced {
			name = src.name
			dstName = dst.name
		} else {
			name = src.namespace
			dstName = dst.namespace
		}

		samples = append(samples, model.Metric{
			kindLabel:       model.LabelValue(name),
			"namespace":     model.LabelValue(src.namespace),
			dstKindLabel:    model.LabelValue(dstName),
			"dst_namespace": model.LabelValue(dst.namespace),
		})
	}

	val, interval := tester.MockQuery(kind, samples, 10)

	var apiPath string

	if pth.Namespaced {
		apiPath = path.Join(
			"/",
			"namespaces",
			edgeRoot.namespace,
			pth.Name,
			edgeRoot.name,
			"edges")
	} else {
		apiPath = path.Join(
			"/",
			"namespaces",
			edgeRoot.namespace,
			"edges")
	}

	log.Info(apiPath)

	rr := s.request("GET", apiPath)

	var resp *metrics.TrafficMetricsList

	assert.NoError(json.Unmarshal(rr.Body.Bytes(), &resp))

	assert.Equal("TrafficMetricsList", resp.TypeMeta.Kind)
	assert.Equal(metrics.APIVersion, resp.TypeMeta.APIVersion)

	assert.Contains(resp.ListMeta.SelfLink, apiPath)

	assert.Equal(kind, resp.Resource.Kind)
	if pth.Namespaced {
		assert.Equal(edgeRoot.namespace, resp.Resource.Namespace)
		assert.Equal(edgeRoot.name, resp.Resource.Name)
	} else {
		assert.Equal(edgeRoot.namespace, resp.Resource.Name)
		assert.Empty(resp.Resource.Namespace)
	}

	require.Len(resp.Items, len(edgeData))

	for i, sample := range edgeData {
		sample := sample

		s.validateTrafficMetrics(
			kind, val, interval, edgeRoot, resp.Items[i])
		s.validateEdge(kind, sample, resp.Items[i], apiPath)
	}

}

func (s *EdgeTestSuite) TestGet() {
	for kind := range metrics.AvailableKinds {
		kind := kind

		s.Run(kind, func() {
			s.testCase(kind)
		})
	}
}

func TestEdge(t *testing.T) {
	suite.Run(t, new(EdgeTestSuite))
}
