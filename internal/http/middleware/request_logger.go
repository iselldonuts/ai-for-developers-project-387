package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	if r.statusCode != 0 {
		return
	}

	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}

	return r.ResponseWriter.Write(body)
}

func RequestLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			recorder := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r)

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", recorder.statusCode).
				Dur("duration", time.Since(startedAt)).
				Msg("request completed")
		})
	}
}
