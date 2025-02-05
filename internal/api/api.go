package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clintjedwards/rc3/internal/conf"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type RouteEntry struct {
	Pattern string
	Router  func(r chi.Router)
}

func startServer(conf *conf.API, routes ...RouteEntry) {
	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Auto-generate a request ID for us.
	router.Use(middleware.RealIP)    // Automatically insert the correct external IP.
	router.Use(middleware.Recoverer) // Don't let panics bring down the entire service.

	for _, route := range routes {
		router.Route(route.Pattern, route.Router)
	}

	httpServer := http.Server{
		Addr:         conf.Server.Host,
		Handler:      loggingMiddleware(router),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine and listen for signals that indicate graceful shutdown
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server exited abnormally")
		}
	}()
	log.Info().Str("url", conf.Server.Host).Msg("started RC3 REST API service")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c

	// Doesn't block if no connections, otherwise will wait until the timeout deadline or connections to finish,
	// whichever comes first.
	ctx, cancel := context.WithTimeout(context.Background(), conf.Server.ShutdownTimeout) // shutdown gracefully
	defer cancel()

	err := httpServer.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not shutdown server in timeout specified")
		return
	}

	log.Info().Msg("server exited gracefully")
}

func StartAPIServer(conf *conf.API) {
	startServer(conf,
		instancesRouter(), // /instances
	)
}

// The logging middleware has to be run before the final call to return the request.
// This is because we wrap the responseWriter to gain information from it after it
// has been written to.
// To speed this process up we call Serve as soon as possible and log afterwards.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		log.Debug().Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status_code", ww.Status()).
			Int("response_size_bytes", ww.BytesWritten()).
			Dur("elapsed_ms", time.Since(start)).
			Msg("")
	})
}
