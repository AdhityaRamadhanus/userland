package mailing

import (
	"net/http"
	"os"
	"time"

	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/NYTimes/gziphandler"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

//Server hold mux Router and information of host port and address of our app
type Server struct {
	Router *mux.Router
	Port   string
}

//NewServer create Server from Handler
func NewServer(Handlers []Handler) *Server {
	router := mux.NewRouter().
		StrictSlash(true).
		PathPrefix("/api").
		Subrouter()

	for _, handler := range Handlers {
		handler.RegisterRoutes(router)
	}

	return &Server{
		Router: router,
		Port:   os.Getenv("USERLAND_MAIL_PORT"),
	}
}

//CreateHttpServer will return http.Server for flexible use like testing
func (s *Server) CreateHttpServer() *http.Server {
	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "mailing",
		Subsystem: "server",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, middlewares.LogMetricKeys)

	summary := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "mailing",
		Subsystem: "server",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, middlewares.LogMetricKeys)

	middlewares := []alice.Constructor{
		middlewares.PanicHandler,
		gziphandler.GzipHandler,
		middlewares.TraceRequest,
		middlewares.LogMetricRequest(counter, summary),
	}
	srv := &http.Server{
		Handler:      alice.New(middlewares...).Then(s.Router),
		Addr:         ":" + os.Getenv("USERLAND_MAIL_PORT"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	return srv
}
