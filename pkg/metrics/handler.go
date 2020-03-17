package metrics

import (
	"net/http"

	"github.com/deislabs/smi-metrics/pkg/mesh"

	metrics "github.com/deislabs/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	"github.com/go-chi/chi"
	"github.com/unrolled/render"
)

// Handler provides the routes required to serve TrafficMetrics
type Handler struct {
	Mesh mesh.Mesh

	render *render.Render
}

// NewHandler returns a handler that has been initialized to defaults.
func NewHandler(meshInstance mesh.Mesh) (*Handler, error) {
	return &Handler{
		Mesh:   meshInstance,
		render: render.New(),
	}, nil
}

// Resources returns all the supported resource types for this server
func (h *Handler) resources(w http.ResponseWriter, r *http.Request) {

	lst, err := h.Mesh.GetSupportedResources(r.Context())
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	h.jsonResponse(w, http.StatusOK, lst)
}

// List returns a list of TrafficMetrics for a specific resource type
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	kind := r.Context().Value(detailsKey).(*mesh.ResourceDetails).Kind
	interval := r.Context().Value(intervalKey).(*metrics.Interval)

	namespace := chi.URLParam(r, "namespace")

	query := mesh.Query{
		Name:      "",
		Namespace: namespace,
		Kind:      kind,
	}

	resourceMetrics, err := h.Mesh.GetResourceMetrics(r.Context(), query, interval)

	if err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, resourceMetrics)
}

// Get a set of metrics for a specific resource
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	interval := r.Context().Value(intervalKey).(*metrics.Interval)
	kind := r.Context().Value(detailsKey).(*mesh.ResourceDetails).Kind

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	query := mesh.Query{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	}

	resourceMetrics, err := h.Mesh.GetResourceMetrics(r.Context(), query, interval)
	if err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(resourceMetrics.Items) == 1 {
		h.jsonResponse(w, http.StatusOK, resourceMetrics.Items[0])
	} else {
		h.jsonResponse(w, http.StatusOK, resourceMetrics)
	}
}

func (h *Handler) edges(w http.ResponseWriter, r *http.Request) {
	interval := r.Context().Value(intervalKey).(*metrics.Interval)
	details := r.Context().Value(detailsKey).(*mesh.ResourceDetails)
	kind := details.Kind

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	query := mesh.Query{
		Name:      name,
		Namespace: namespace,
		Kind:      kind,
	}

	edgeMetrics, err := h.Mesh.GetEdgeMetrics(r.Context(), query, interval, details)
	if err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, edgeMetrics)

}
