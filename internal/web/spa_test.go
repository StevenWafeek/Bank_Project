package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func get(h http.Handler, url string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	return rec
}

func TestSPA(t *testing.T) {
	dir := t.TempDir()
	if rec := get(SPA(dir), "/"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing build: expected 503, got %d", rec.Code)
	}
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<div id=root></div>"), 0o644)
	os.MkdirAll(filepath.Join(dir, "assets"), 0o755)
	os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log(1)"), 0o644)
	h := SPA(dir)

	for _, url := range []string{"/", "/dashboard", "/about"} {
		if rec := get(h, url); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "root") {
			t.Errorf("%s: expected index.html, got %d %q", url, rec.Code, rec.Body.String())
		}
	}
	if rec := get(h, "/assets/app.js"); !strings.Contains(rec.Body.String(), "console.log") {
		t.Errorf("asset not served: %q", rec.Body.String())
	}
}
