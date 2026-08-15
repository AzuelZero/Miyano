// Package handler holds the HTTP handlers of the API.
package handler

import (
	"encoding/json"
	"net/http"
)

// Health returns the healthcheck handler. It signals liveness to the
// hosting platform and to humans poking at the API.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
