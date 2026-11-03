package middleware

import "net/http"

// BodyLimit caps request bodies at max bytes by wrapping r.Body in an
// http.MaxBytesReader. JSON decoding then fails with *http.MaxBytesError,
// which the handlers map to 413 (audit 03/11: unbounded bodies fed argon2id
// and routine inserts a CPU/RAM/connection-pool lever).
func BodyLimit(max int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, max)
			next.ServeHTTP(w, r)
		})
	}
}
