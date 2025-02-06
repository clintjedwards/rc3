package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *APIContext) instancesRouter() RouteEntry {
	router := func(router chi.Router) {
		router.Get("/", api.getInstances)
	}

	return RouteEntry{
		Pattern: "/instances",
		Router:  router,
	}
}

func (api *APIContext) getInstances(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!"))
}
