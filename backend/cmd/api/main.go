// Package main es el punto de entrada de la API de Miyano.
// Por ahora arranca un servidor HTTP mínimo con un endpoint /health.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// puertoPorDefecto es el puerto en el que escucha la API si no se define PORT.
const puertoPorDefecto = "8080"

func main() {
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = puertoPorDefecto
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

	srv := &http.Server{
		Addr:         ":" + puerto,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Miyano API escuchando en :%s", puerto)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("error arrancando el servidor: %v", err)
	}
}

// health responde con un JSON sencillo para que el hosting (y tú) comprueben
// que la API está viva. Es el endpoint canónico de healthcheck.
func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{
		"status": "ok",
	}
	_ = json.NewEncoder(w).Encode(resp)
}
