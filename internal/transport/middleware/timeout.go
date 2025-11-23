package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Timeout устанавливает максимальное время выполнения запроса
func Timeout(timeout time.Duration, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				defer close(done)
				next.ServeHTTP(w, r)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				logger.Warn("request timeout",
					zap.String("request_id", middleware.GetReqID(r.Context())),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Duration("timeout", timeout),
				)
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
			}
		})
	}
}

