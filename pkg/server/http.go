package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
)

func RunHTTPServer(createHandler func(router chi.Router) http.Handler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	RunHTTPServerOnAddr(":"+port, createHandler)
}

func RunHTTPServerOnAddr(addr string, createHandler func(router chi.Router) http.Handler) {
	apiRouter := chi.NewRouter()

	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.RealIP)
	apiRouter.Use(middleware.Logger)
	apiRouter.Use(middleware.Recoverer)

	addCorsMiddleware(apiRouter)

	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", createHandler(apiRouter))

	logrus.Info("Starting HTTP server at: %s.", addr)
	if err := http.ListenAndServe(addr, rootRouter); err != nil {
		logrus.WithError(err).Panic("Unable to start HTTP server.")
	}
}

func addCorsMiddleware(router *chi.Mux) {
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ";")
	if len(allowedOrigins) == 0 {
		return
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	router.Use(corsMiddleware.Handler)
}
