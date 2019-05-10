package metrics

import (
	"github.com/go-chi/chi"
)

// Routes returns the routes that allow fetching metrics
func (h *Handler) Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", h.resources)

	router.Group(func(router chi.Router) {
		router.Use(h.addResourceDetails)
		router.Use(h.addInterval)

		router.Get("/{apiResourceName}", h.list)
		// This will only every match namespaces, but it allows the same API to
		// work for non-namespaced resources
		router.Get("/{apiResourceName:namespaces}/{name}", h.get)
		router.Get("/{apiResourceName:namespaces}/{name}/edges", h.edges)

		router.Get("/namespaces/{namespace}/{apiResourceName:[^e].*}", h.list)

		router.Get(
			"/namespaces/{namespace}/{apiResourceName}/{name}", h.get)

		router.Get(
			"/namespaces/{namespace}/{apiResourceName}/{name}/edges",
			h.edges)
	})

	return router
}
