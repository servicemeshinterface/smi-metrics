package linkerd

import (
	"encoding/json"
	"testing"

	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HandlerTestSuite struct {
	Suite
}

func (s *HandlerTestSuite) TestResources() {
	assert := s.Assert()

	rr := s.request("GET", "/")

	var resp *metav1.APIResourceList

	assert.NoError(json.Unmarshal(rr.Body.Bytes(), &resp))

	assert.Equal("APIResourceList", resp.TypeMeta.Kind)
	assert.Equal(metrics.APIVersion, resp.GroupVersion)

	assert.Len(resp.APIResources, len(metrics.AvailableKinds))
	for _, res := range resp.APIResources {
		assert.Equal("TrafficMetrics", res.Kind)
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
