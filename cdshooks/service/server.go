package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	cdshooks "github.com/your-org/cds-hooks-go/cdshooks"
)

type Server struct {
	registry        map[string]cdshooks.ServiceEntry
	logger          *slog.Logger
	corsOrigins     []string
	requestTimeout  time.Duration
	feedbackHandler FeedbackHandler
	mu              sync.RWMutex
}

type ServerOption func(*Server)

func WithLogger(l *slog.Logger) ServerOption {
	return func(s *Server) {
		s.logger = l
	}
}

func WithCORSOrigins(origins ...string) ServerOption {
	return func(s *Server) {
		s.corsOrigins = origins
	}
}

func WithRequestTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.requestTimeout = d
	}
}

func WithFeedbackHandler(h FeedbackHandler) ServerOption {
	return func(s *Server) {
		s.feedbackHandler = h
	}
}

func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		registry: make(map[string]cdshooks.ServiceEntry),
		logger:   slog.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Register(entries ...cdshooks.ServiceEntry) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range entries {
		if entry.Service.ID == "" {
			continue
		}
		s.registry[entry.Service.ID] = entry
	}

	return s
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.loggingMiddleware)
	r.Use(s.recoveryMiddleware)
	r.Use(s.corsMiddleware)

	r.Get("/cds-services", s.handleDiscovery)
	r.Post("/cds-services/{id}", s.handleService)
	if s.feedbackHandler != nil {
		r.Post("/cds-services/{id}/feedback", s.handleFeedback)
	}

	return r
}

func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}

	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("server starting", "addr", addr)
		errChan <- srv.ListenAndServe()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		s.logger.Info("shutting down server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		s.logger.Info("request",
			"method", method,
			"path", path,
			"status", wrapper.statusCode,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.Error("panic recovered", "panic", rec)
				s.writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		hasOrigin := origin != ""
		allowed := len(s.corsOrigins) == 0

		if !allowed && hasOrigin {
			for _, o := range s.corsOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}
		}

		if allowed && hasOrigin {
			if len(s.corsOrigins) == 1 && s.corsOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]cdshooks.Service, 0, len(s.registry))
	for _, entry := range s.registry {
		services = append(services, entry.Service)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]cdshooks.Service{
		"services": services,
	})
}

func (s *Server) handleService(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")

	s.mu.RLock()
	entry, exists := s.registry[serviceID]
	s.mu.RUnlock()

	if !exists {
		s.writeError(w, http.StatusNotFound, "service not found")
		return
	}

	var req cdshooks.CDSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.HookInstance == "" {
		s.writeError(w, http.StatusBadRequest, "hookInstance is required")
		return
	}

	if _, err := uuid.Parse(req.HookInstance); err != nil {
		s.writeError(w, http.StatusBadRequest, "hookInstance must be a valid UUID")
		return
	}

	ctx := r.Context()
	if s.requestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.requestTimeout)
		defer cancel()
	}

	resp, err := entry.Handler.Handle(ctx, req)
	if err != nil {
		s.logger.Error("handler error", "error", err)
		s.writeError(w, http.StatusInternalServerError, "handler error")
		return
	}

	if resp.Cards == nil {
		resp.Cards = []cdshooks.Card{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	if s.feedbackHandler == nil {
		s.writeError(w, http.StatusNotFound, "feedback not enabled")
		return
	}

	serviceID := chi.URLParam(r, "id")

	var feedback cdshooks.FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()
	if err := s.feedbackHandler.Feedback(ctx, serviceID, feedback); err != nil {
		s.logger.Error("feedback error", "error", err)
		s.writeError(w, http.StatusInternalServerError, "feedback error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cdshooks.FeedbackResponse{Status: "ok"})
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type FeedbackHandler interface {
	Feedback(ctx context.Context, serviceID string, feedback cdshooks.FeedbackRequest) error
}
