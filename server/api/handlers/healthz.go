package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HealthzHandler struct{}

func (h HealthzHandler) RegisterRoutes(router *mux.Router) {
	healthz := http.HandlerFunc(h.healthz)
	router.HandleFunc("/", healthz).Methods("GET")
	router.HandleFunc("/healthz", healthz).Methods("GET")
}

func (h *HealthzHandler) healthz(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}
