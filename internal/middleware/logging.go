package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/dagger021/log-ingestion-query-system/internal/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Logger(logger *zap.Logger) func(nextā http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.New()

			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			reqLogger := logger.With(
				zap.String("request_id", requestID.String()),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Time("started_at", start),
			)

			ctx := context.WithValue(r.Context(), config.LoggerKey, reqLogger)
			next.ServeHTTP(rec, r.WithContext(ctx))

			// logging request based on severity of status code
			loggerFunc := reqLogger.Info
			switch {
			case rec.status >= 500:
				loggerFunc = reqLogger.Error
			case rec.status >= 400:
				loggerFunc = reqLogger.Warn
			}

			loggerFunc("request completed",
				zap.Int("status", rec.status),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
