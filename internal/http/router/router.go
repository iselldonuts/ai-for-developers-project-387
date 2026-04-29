package router

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	httpapi "github.com/iselldonuts/ai-for-developers-project-386/internal/http/api"
)

func New(healthHandler http.Handler, apiHandler httpapi.ServerInterface, middlewares ...func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthHandler)
	mux.Handle("/", spaHandler("web/dist"))

	var handler http.Handler = httpapi.HandlerFromMux(apiHandler, mux)

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func spaHandler(root string) http.Handler {
	files := http.Dir(root)
	fileServer := http.FileServer(files)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(filepath.Clean(r.URL.Path), string(filepath.Separator))
		if path == "." {
			path = "index.html"
		}

		if hasStaticExtension(path) {
			fileServer.ServeHTTP(w, r)
			return
		}

		if exists(files, path) {
			fileServer.ServeHTTP(w, r)
			return
		}

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = cloneURL(r.URL)
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	})
}

func exists(files http.FileSystem, name string) bool {
	file, err := files.Open(name)
	if err != nil {
		return false
	}
	_ = file.Close()

	return true
}

func hasStaticExtension(path string) bool {
	return filepath.Ext(path) != ""
}

func cloneURL(url *url.URL) *url.URL {
	copied := *url
	return &copied
}
