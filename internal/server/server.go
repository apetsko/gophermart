package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func New(addr string, r *chi.Mux) *http.Server {
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           r,
	}
	return server
}
