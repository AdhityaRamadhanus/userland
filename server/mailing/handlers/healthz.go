package handlers

import (
	"net/http"

	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/gorilla/mux"
)

type HealthzHandler struct{}

func (h HealthzHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/healthz", middlewares.BasicAuth("test", "coba", h.healthz)).Methods("GET")
}

func (h *HealthzHandler) healthz(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}
