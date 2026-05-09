package main

import (
	"context"
	"net/http"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Server struct {
	server *http.Server
	cfg    *config
	logger *zap.Logger
}

func NewServer(cfg *config, logger *zap.Logger, sentryHandler *sentryhttp.Handler) Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/cv", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/cv", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=\"cv.pdf\"")
		http.ServeFile(w, r, "static/cv.pdf")
	})

	s := &http.Server{
		Addr:    cfg.Addr(),
		Handler: recoveryMiddleware(logger, loggerMiddleware(logger, reqIdMiddleware(sentryHandler.Handle(mux)))),
	}

	return Server{server: s, cfg: cfg, logger: logger}
}

func (s *Server) Start() error {
	s.logger.Info("server started", zap.String("addr", s.cfg.Addr()))
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("server stopping")
	return s.server.Shutdown(ctx)
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggerMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.status),
			zap.Duration("duration", time.Since(start)),
			zap.String("user_agent", r.UserAgent()),
		}
		if reqID := r.Header.Get("X-Request-Id"); reqID != "" {
			fields = append(fields, zap.String("request_id", reqID))
		}

		logger.Info("http request", fields...)
	})
}

func reqIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = generateReqID()
			r.Header.Set("X-Request-Id", reqID)
		}
		next.ServeHTTP(w, r)
	})
}

func generateReqID() string {
	return uuid.New().String()
}

func recoveryMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", r.URL.Path),
					zap.Stack("stacktrace"),
				)
				panic(err) // re-panic so sentryhttp can catch and report it
			}
		}()
		next.ServeHTTP(w, r)
	})
}
