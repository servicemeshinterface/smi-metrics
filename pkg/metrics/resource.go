package metrics

import (
	"net/http"
	"strings"

	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	"github.com/go-chi/chi"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/deislabs/smi-metrics/pkg/prometheus"
)

type resourceLookup struct {
	Item     *metrics.TrafficMetricsList
	interval *metrics.Interval
}

func (r *resourceLookup) Get(labels model.Metric) *metrics.TrafficMetrics {
	labelName := model.LabelName(strings.ToLower(r.Item.Resource.Kind))

	obj := r.Item.Get(listKey(
		r.Item.Resource.Kind,
		string(labels[labelName]),
		string(labels["namespace"]),
	), nil)
	obj.Interval = r.interval
	obj.Edge = &metrics.Edge{
		Direction: metrics.From,
	}

	return obj
}

func (r *resourceLookup) Queries() []*prometheus.Query {
	queries := []*prometheus.Query{}
	for name, tmpl := range resourceQueries {
		queries = append(queries, &prometheus.Query{
			Name:     name,
			Template: tmpl,
			Values: map[string]interface{}{
				"kind":      r.Item.Resource.Kind,
				"namespace": r.Item.Resource.Namespace,
				"name":      r.Item.Resource.Name,
			},
		})
	}

	return queries
}

// List returns a list of TrafficMetrics for a specific resource type
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	kind := r.Context().Value(detailsKey).(*resourceDetails).Kind
	interval := r.Context().Value(intervalKey).(*metrics.Interval)

	namespace := chi.URLParam(r, "namespace")

	lookup := &resourceLookup{
		Item: metrics.NewTrafficMetricsList(&v1.ObjectReference{
			Kind: kind,
			// If a namespace isn't defined, it'll be the empty string which fits
			// with the struct's idea of "empty"
			Namespace: namespace,
		}, false),
		interval: interval,
	}

	if err := prometheus.NewClient(r.Context(), h.client, interval).Update(
		lookup); err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, errorResponse{
			Error: "unable to lookup metrics",
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, lookup.Item)
}

// Get a set of metrics for a specific resource
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	interval := r.Context().Value(intervalKey).(*metrics.Interval)
	kind := r.Context().Value(detailsKey).(*resourceDetails).Kind

	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	// Get is somewhat of a special case as *most* handlers just return a list.
	// Create a list with a fully specified object reference and then just
	// return a single element to keep the code as similar as possible.
	lookup := &resourceLookup{
		Item: metrics.NewTrafficMetricsList(&v1.ObjectReference{
			Kind: kind,
			Name: name,
			// If a namespace isn't defined, it'll be the empty string which fits
			// with the struct's idea of "empty"
			Namespace: namespace,
		}, false),
		interval: interval,
	}

	if err := prometheus.NewClient(r.Context(), h.client, interval).Update(
		lookup); err != nil {
		log.Error(err)
		h.jsonResponse(w, http.StatusInternalServerError, errorResponse{
			Error: "unable to lookup metrics",
		})
		return
	}

	if len(lookup.Item.Items) != 1 {
		for _, x := range lookup.Item.Items {
			log.Info(x.Resource)
		}
		log.Errorf("Wrong number of items: %d", len(lookup.Item.Items))
		h.jsonResponse(w, http.StatusInternalServerError, errorResponse{
			Error: "unable to lookup metrics",
		})
		return
	}

	h.jsonResponse(w, http.StatusOK, lookup.Item.Items[0])
}
