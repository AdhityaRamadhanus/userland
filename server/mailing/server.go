package mailing

import (
	"net/http"
	"os"
	"time"

	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/common/metrics"
	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

//Server hold mux Router and information of host port and address of our app
type Server struct {
	Router *mux.Router
	Port   string
}

//NewServer create Server from Handler
func NewServer(Handlers []Handler) *Server {
	router := mux.NewRouter().StrictSlash(true).

	for _, handler := range Handlers {
		handler.RegisterRoutes(router)
	}

	return &Server{
		Router: router,
		Port:   os.Getenv("USERLAND_MAIL_PORT"),
	}
}

//CreateHTTPServer will return http.Server for flexible use like testing
func (s Server) CreateHTTPServer() *http.Server {
	middlewares := []alice.Constructor{
		middlewares.PanicHandler,
		gziphandler.GzipHandler,
		middlewares.TraceRequest,
		alice.Constructor(middlewares.LogMetricRequest(
			metrics.PrometheusRequestCounter("mailing", "server", middlewares.LogMetricKeys),
			metrics.PrometheusRequestLatency("mailing", "server", middlewares.LogMetricKeys),
		)),
	}
	srv := &http.Server{
		Handler:      alice.New(middlewares...).Then(s.Router),
		Addr:         ":" + os.Getenv("USERLAND_MAIL_PORT"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	return srv
}
