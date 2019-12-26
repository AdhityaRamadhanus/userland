package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/cors"
)

//Server hold mux Router and information of host port and address of our app
type Server struct {
	Router  *mux.Router
	Address string
}

//NewServer create Server from Handler
func NewServer(cfg config.ApiConfig, Handlers ...Handler) *Server {
	router := mux.NewRouter().StrictSlash(true)

	for _, handler := range Handlers {
		handler.RegisterRoutes(router)
	}

	return &Server{
		Router:  router,
		Address: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

//CreateHTTPServer will return http.Server for flexible use like testing
func (s Server) CreateHTTPServer() *http.Server {
	middlewares := []alice.Constructor{
		middlewares.PanicHandler,
		gziphandler.GzipHandler,
		middlewares.TraceRequest,
		middlewares.ParseClientInfo,
		cors.Default().Handler,
		alice.Constructor(middlewares.LogMetricRequest(
			metrics.PrometheusRequestCounter("api", "server", middlewares.LogMetricKeys),
			metrics.PrometheusRequestLatency("api", "server", middlewares.LogMetricKeys),
		)),
	}

	srv := &http.Server{
		Handler:      alice.New(middlewares...).Then(s.Router),
		Addr:         s.Address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	return srv
}
