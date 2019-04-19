package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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

type resourceDetails struct {
	Kind       string
	Namespaced bool
}

func getResourceDetails(name string) (*resourceDetails, bool) {
	for kind, r := range metrics.AvailableKinds {
		if r.Name == name {
			return &resourceDetails{
				Kind:       kind,
				Namespaced: r.Namespaced,
			}, true
		}
	}

	return nil, false
}

// AddResourceDefinition adds a api resource definition to context for use in
// handlers
func (h *Handler) addResourceDetails(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiResourceName := chi.URLParam(r, "apiResourceName")

		details, ok := getResourceDetails(apiResourceName)
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

// AddInterval adds an interval to the context for use in handlers
func (h *Handler) addInterval(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := context.WithValue(r.Context(), intervalKey, &metrics.Interval{
			Timestamp: metav1.NewTime(start),
			Window:    metav1.Duration{Duration: 30 * time.Second},
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// listKey constructs a key that is useful for getting elements from a
// TrafficMetricsList
func listKey(kind, name, namespace string) *v1.ObjectReference {
	var namespaced bool

	if details, ok := metrics.AvailableKinds[kind]; ok {
		namespaced = details.Namespaced
	} else {
		namespaced = true
	}

	obj := &v1.ObjectReference{
		Kind: kind,
		Name: name,
	}

	if namespaced {
		obj.Namespace = namespace
	}

	return obj
}
