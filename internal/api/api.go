package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clintjedwards/rc3/internal/conf"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luthermonson/go-proxmox"
	"github.com/rs/zerolog/log"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r)
	})
}

// Data kept for the lifetime of the API.
type APIContext struct {
	Client        *proxmox.Client
	ProxmoxConfig *conf.Proxmox
}

func newAPIContext(proxmoxConf *conf.Proxmox) *APIContext {
	var client *proxmox.Client

	if proxmoxConf.UseTLS {
		client = proxmox.NewClient(proxmoxConf.URL, proxmox.WithAPIToken(proxmoxConf.TokenID, proxmoxConf.TokenSecret))
	} else {
		insecureHTTPClient := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		client = proxmox.NewClient(proxmoxConf.URL,
			proxmox.WithAPIToken(proxmoxConf.TokenID, proxmoxConf.TokenSecret),
			proxmox.WithHTTPClient(&insecureHTTPClient),
		)
	}

	version, err := client.Version(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("could not successfully connect to proxmox using provided url/credentials")
	}

	log.Info().Str("url", proxmoxConf.URL).
		Bool("tls", proxmoxConf.UseTLS).
		Str("token_id", proxmoxConf.TokenID).
		Str("version", version.Version).
		Msg("successfully connected to proxmox")

	return &APIContext{
		Client:        client,
		ProxmoxConfig: proxmoxConf,
	}
}

type RouteEntry struct {
	Pattern string
	Router  func(r chi.Router)
}

func startServer(conf *conf.API, routes ...RouteEntry) {
	router := chi.NewRouter()

	router.Use(middleware.RequestID) // Auto-generate a request ID for us.
	router.Use(requestIDMiddleware)  // Ensure request ID is attached to every response.
	router.Use(middleware.RealIP)    // Automatically insert the correct external IP.
	router.Use(middleware.Recoverer) // Don't let panics bring down the entire service.
	router.Use(loggingMiddleware)    // Log requests
	router.Route("/api", func(r chi.Router) {
		for _, route := range routes {
			r.Route(route.Pattern, route.Router)
		}
	})

	httpServer := http.Server{
		Addr:         conf.Server.Host,
		Handler:      router,
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
	api := newAPIContext(conf.Proxmox)

	startServer(conf,
		api.instancesRouter(), // /api/instances
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

type AuthContext struct {
	RecurserID string
}

func CheckAuth(r *http.Request) AuthContext {
	return AuthContext{}
}

type ErrorResponse struct {
	Error        string `json:"error"`         // Short description
	ErrorDetails string `json:"error_details"` // More detailed explanation
}

func writeError(w http.ResponseWriter, statusCode int, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:        http.StatusText(statusCode),
		ErrorDetails: details,
	})
}

func writeResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode response from server")
		return
	}
}
