package server

import (
	"net/http"
	"os"
	"time"

	"github.com/AdhityaRamadhanus/userland/server/middlewares"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/cors"
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
		Port:   os.Getenv("USERLAND_PORT"),
	}
}

//CreateHttpServer will return http.Server for flexible use like testing
func (s *Server) CreateHttpServer() *http.Server {
	middlewares := []alice.Constructor{
		middlewares.PanicHandler,
		middlewares.Gzip,
		middlewares.TraceRequest,
		cors.Default().Handler,
		middlewares.LogRequest,
	}
	srv := &http.Server{
		Handler:      alice.New(middlewares...).Then(s.Router),
		Addr:         ":" + os.Getenv("USERLAND_PORT"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	return srv
}
