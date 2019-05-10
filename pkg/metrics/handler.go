package metrics

import (
	"net/http"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/unrolled/render"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Handler provides the routes required to serve TrafficMetrics
type Handler struct {
	client promv1.API

	render *render.Render
}

// NewHandler returns a handler that has been initialized to defaults.
func NewHandler(url, groupVersion string) (*Handler, error) {
	promClient, err := api.NewClient(api.Config{Address: url})
	if err != nil {
		return nil, err
	}

	return &Handler{
		client: promv1.NewAPI(promClient),
		render: render.New(),
	}, nil
}

// Resources returns all the supported resource types for this server
func (h *Handler) resources(w http.ResponseWriter, r *http.Request) {
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

	h.jsonResponse(w, http.StatusOK, lst)
}
