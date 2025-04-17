package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/config"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	userService "github.com/Abdelrahman-habib/expense-tracker/internal/users/service"

	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"go.uber.org/zap"
)

type Middleware struct {
	logger      *zap.Logger
	auth        service.Service
	db          db.Service
	config      config.ServerConfig
	userService userService.UsersService
	cache       interface{}
}

var responseWriterPool = sync.Pool{
	New: func() interface{} {
		return &responseWriter{
			status: http.StatusOK,
		}
	},
}

func NewMiddleware(logger *zap.Logger, auth service.Service, db db.Service, config config.ServerConfig, cache interface{}) *Middleware {
	return &Middleware{
		logger: logger,
		auth:   auth,
		db:     db,
		config: config,
		cache:  cache,
	}
}

// Timeout middleware cancels the context after the specified duration
func (m *Middleware) Timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			tw := &timeoutWriter{
				w:      w,
				logger: m.logger,
			}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				tw.mu.Lock()
				defer tw.mu.Unlock()

				if !tw.written {
					w.WriteHeader(http.StatusGatewayTimeout)
					m.logger.Warn("request timed out",
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.Duration("timeout", timeout),
					)
				}
			}
		})
	}
}

// Logger logs request details
func (m *Middleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := responseWriterPool.Get().(*responseWriter)
		writer.ResponseWriter = w
		writer.status = http.StatusOK
		defer func() {
			writer.ResponseWriter = nil // Clear reference
			responseWriterPool.Put(writer)
		}()

		start := time.Now()
		next.ServeHTTP(writer, r)

		m.logger.Info("request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", writer.status),
			zap.Duration("duration", time.Since(start)),
			zap.String("ip", r.RemoteAddr),
			zap.String("user-agent", r.UserAgent()),
		)
	})
}

// CORS sets up CORS headers
func (m *Middleware) CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   m.config.Middleware.AllowedOrigins,
		AllowedMethods:   m.config.Middleware.AllowedMethods,
		AllowedHeaders:   m.config.Middleware.AllowedHeaders,
		ExposedHeaders:   m.config.Middleware.ExposedHeaders,
		AllowCredentials: m.config.Middleware.AllowCredentials,
		MaxAge:           m.config.Middleware.MaxAge,
	})
}

// RateLimiter implements rate limiting
func (m *Middleware) RateLimiter(next http.Handler) http.Handler {
	return httprate.LimitByIP(
		m.config.Middleware.RateLimit.RequestsPerMinute,
		m.config.Middleware.RateLimit.WindowLength,
	)(next)
}

// Recovery handles panics
func (m *Middleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", r.URL.Path),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// clerk auth
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return m.auth.Middleware(next)
}

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// timeoutWriter wraps http.ResponseWriter to track if headers were written
type timeoutWriter struct {
	w       http.ResponseWriter
	mu      sync.Mutex
	written bool
	logger  *zap.Logger
}

func (tw *timeoutWriter) Header() http.Header {
	return tw.w.Header()
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.written = true
	return tw.w.Write(b)
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.written = true
	tw.w.WriteHeader(code)
}
