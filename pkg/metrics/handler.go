package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deislabs/smi-metrics/pkg/mesh"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
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

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	if err := h.render.JSON(w, status, payload); err != nil {
		log.Errorf("error rendering response: %s", err)
		http.Error(w, "error rendering response", 500)
	}
}

// AddResourceDefinition adds a api resource definition to context for use in
// handlers
func (h *Handler) addResourceDetails(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiResourceName := chi.URLParam(r, "apiResourceName")

		details, ok := mesh.GetResourceDetails(apiResourceName)
		if !ok {
			h.jsonResponse(w, http.StatusNotFound, map[string]string{
				"error": fmt.Sprintf("Unsupported resource: %s", apiResourceName),
			})
			return
		}

		ctx := context.WithValue(r.Context(), mesh.DetailsKey, details)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AddInterval adds an interval to the context for use in handlers
func (h *Handler) addInterval(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := context.WithValue(r.Context(), mesh.IntervalKey, &metrics.Interval{
			Timestamp: metav1.NewTime(start),
			Window:    metav1.Duration{Duration: 30 * time.Second},
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Resources returns all the supported resource types for this server
func (h *Handler) resources(w http.ResponseWriter, r *http.Request) {

	lst, err := h.Mesh.GetSupportedResources(r.Context())
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, err.Error())
	}
	h.jsonResponse(w, http.StatusOK, lst)
}

// List returns a list of TrafficMetrics for a specific resource type
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	kind := r.Context().Value(mesh.DetailsKey).(*mesh.ResourceDetails).Kind
	interval := r.Context().Value(mesh.IntervalKey).(*metrics.Interval)

	namespace := chi.URLParam(r, "namespace")

	resourceMetrics, err := h.Mesh.GetResourceMetrics(r.Context(), "", namespace, kind, interval)

	if err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, resourceMetrics)
}

// Get a set of metrics for a specific resource
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	interval := r.Context().Value(mesh.IntervalKey).(*metrics.Interval)
	kind := r.Context().Value(mesh.DetailsKey).(*mesh.ResourceDetails).Kind

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	resourceMetrics, err := h.Mesh.GetResourceMetrics(r.Context(), name, namespace, kind, interval)
	if err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(resourceMetrics.Items) != 1 {
		for _, x := range resourceMetrics.Items {
			log.Info(x.Resource)
		}
		log.Errorf("Wrong number of items: %d", len(resourceMetrics.Items))
		h.jsonResponse(w, http.StatusInternalServerError, mesh.ErrorResponse{
			Error: "unable to lookup metrics",
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, resourceMetrics.Items[0])
}

func (h *Handler) edges(w http.ResponseWriter, r *http.Request) {
	interval := r.Context().Value(mesh.IntervalKey).(*metrics.Interval)
	details := r.Context().Value(mesh.DetailsKey).(*mesh.ResourceDetails)
	kind := details.Kind

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	edgeMetrics, err := h.Mesh.GetEdgeMetrics(r.Context(), name, namespace, kind, interval, details)
	if err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, edgeMetrics)

}
