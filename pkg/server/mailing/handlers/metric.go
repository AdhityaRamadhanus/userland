package handlers

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricHandler struct{}

func (h MetricHandler) RegisterRoutes(router *mux.Router) {
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
}
