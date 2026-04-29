package middleware

import (
	"net/http"
	"strings"
)

var (
	allowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions}
	allowedHeaders = []string{"Content-Type"}
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowedOriginsSet := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowedOriginsSet[origin] = struct{}{}
	}

	allowMethodsValue := strings.Join(allowedMethods, ", ")
	allowHeadersValue := strings.Join(allowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowedOriginsSet[origin]; !ok {
				if r.Method == http.MethodOptions {
					http.Error(w, "CORS origin not allowed", http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Vary", "Origin")
			headers.Set("Access-Control-Allow-Methods", allowMethodsValue)
			headers.Set("Access-Control-Allow-Headers", allowHeadersValue)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
