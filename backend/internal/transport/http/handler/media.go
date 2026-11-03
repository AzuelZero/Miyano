package handler

import (
	"net/http"
	"path"
	"strings"
)

// allowedMediaPrefixes is the subset of the dataset clone the app consumes
// (image_url/gif_url are "images/..." and "videos/..." paths). Everything
// else in the clone — repo metadata, HTML pages, data files — stays unserved.
var allowedMediaPrefixes = []string{"images/", "videos/"}

// ServeMedia serves the exercise-dataset clone at /media/* behind an
// allowlist: only the images/ and videos/ subtrees, no directory listings,
// no dotfiles, and X-Content-Type-Options: nosniff on every answer
// (audit 03/11: the bare FileServer exposed .git/, index.html and listings).
func ServeMedia(dir string) http.HandlerFunc {
	fs := http.StripPrefix("/media/", http.FileServer(http.Dir(dir)))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")

		rel := strings.TrimPrefix(r.URL.Path, "/media/")
		if rel == "" || strings.HasSuffix(rel, "/") {
			http.NotFound(w, r) // no directory listing
			return
		}
		// Explicit dot-dot rejection: stdlib already roots http.Dir, this
		// makes the boundary testable instead of implicit.
		if strings.Contains(rel, "..") {
			http.NotFound(w, r)
			return
		}
		allowed := false
		for _, prefix := range allowedMediaPrefixes {
			if strings.HasPrefix(rel, prefix) {
				allowed = true
				break
			}
		}
		if !allowed || strings.HasPrefix(path.Base(rel), ".") {
			http.NotFound(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	}
}
