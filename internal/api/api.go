package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouteEntry struct {
	Pattern string
	Router  func(r chi.Router)
}

func startAPIServer(routes ...RouteEntry) {
	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Auto-generate a request ID for us.
	router.Use(middleware.RealIP)    // Automatically insert the correct external IP.
	router.Use(middleware.Recoverer) // Don't let panics bring down the entire service.

	for _, route := range routes {
		router.Route(route.Pattern, route.Router)
	}

	http.ListenAndServe(":3000", router)
}
