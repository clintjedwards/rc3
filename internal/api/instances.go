package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func instancesRouter() RouteEntry {
	router := func(router chi.Router) {
		router.Get("/", getInstances)
	}

	return RouteEntry{
		Pattern: "/instances",
		Router:  router,
	}
}

func getInstances(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!"))
}
