package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/servicemeshinterface/smi-metrics/pkg/mesh"
	metrics "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type contextKey string

var (
	intervalKey = contextKey("metrics-interval")
	detailsKey  = contextKey("metrics-details")
)

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	if err := h.render.JSON(w, status, payload); err != nil {
		log.Errorf("error rendering response: %s", err)
		http.Error(w, "error rendering response", 500)
	}
}

// addResourceDefinition adds a api resource definition to context for use in
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

		ctx := context.WithValue(r.Context(), detailsKey, details)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// addInterval adds an interval to the context for use in handlers
func (h *Handler) addInterval(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		t := r.URL.Query().Get("t")
		window, err := time.ParseDuration(t)
		if err != nil {
			window = 30 * time.Second
		}

		ctx := context.WithValue(r.Context(), intervalKey, &metrics.Interval{
			Timestamp: metav1.NewTime(start),
			Window:    metav1.Duration{Duration: window},
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
