// Package web serves the built React single-page app.
package web

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// SPA serves static files from dir and falls back to index.html so client-side
// routes like /dashboard work on reload.
func SPA(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(index); err != nil {
			http.Error(w, "Frontend not built. Run: cd frontend && npm install && npm run build", http.StatusServiceUnavailable)
			return
		}
		p := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+r.URL.Path)))
		if info, err := os.Stat(p); err != nil || info.IsDir() {
			http.ServeFile(w, r, index)
			return
		}
		files.ServeHTTP(w, r)
	})
}
