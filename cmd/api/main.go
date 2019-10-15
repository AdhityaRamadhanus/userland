package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	server "github.com/AdhityaRamadhanus/userland/server/api"
	"github.com/AdhityaRamadhanus/userland/server/api/handlers"
	"github.com/AdhityaRamadhanus/userland/service/authentication"
	"github.com/AdhityaRamadhanus/userland/service/event"
	"github.com/AdhityaRamadhanus/userland/service/profile"
	"github.com/AdhityaRamadhanus/userland/service/session"
	"github.com/AdhityaRamadhanus/userland/storage/gcs"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	_redis "github.com/go-redis/redis"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sarulabs/di"
	log "github.com/sirupsen/logrus"
)

func buildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		postgres.ConnectionBuilder,
		redis.ConnectionBuilder("redis-connection", 0),
		redis.ConnectionBuilder("redis-rate-limit-connection", 2),
		gcs.ServiceBuilder,
		mailing.ClientBuilder,
		redis.KeyValueServiceBuilder,
		postgres.UserRepositoryBuilder,
		postgres.EventRepositoryBuilder,
		redis.SessionRepositoryBuilder,
		authentication.ServiceBuilder,
		authentication.ServiceInstrumentorBuilder,
		session.ServiceBuilder,
		session.ServiceInstrumentorBuilder,
		profile.ServiceBuilder,
		profile.ServiceInstrumentorBuilder,
		event.ServiceBuilder,
		event.ServiceInstrumentorBuilder,
	)

	return builder.Build()
}

func init() {
	godotenv.Load()
	switch os.Getenv("ENV") {
	case "production":
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.DebugLevel)
		log.SetOutput(os.Stdout)
	}
}

func main() {
	ctn := buildContainer()

	authenticator := middlewares.NewAuthenticator(ctn.Get("keyvalue-service").(userland.KeyValueService))
	ratelimiter := middlewares.NewRateLimiter(ctn.Get("redis-rate-limit-connection").(*_redis.Client))

	healthHandler := handlers.HealthzHandler{}
	metricHandler := handlers.MetricHandler{}
	authenticationHandler := handlers.AuthenticationHandler{
		RateLimiter:           ratelimiter,
		Authenticator:         authenticator,
		ProfileService:        ctn.Get("profile-service").(profile.Service),
		AuthenticationService: ctn.Get("authentication-service").(authentication.Service),
		SessionService:        ctn.Get("session-service").(session.Service),
		EventService:          ctn.Get("event-service").(event.Service),
	}
	profileHandler := handlers.ProfileHandler{
		RateLimiter:    ratelimiter,
		Authenticator:  authenticator,
		ProfileService: ctn.Get("profile-service").(profile.Service),
		EventService:   ctn.Get("event-service").(event.Service),
	}
	sessionHandler := handlers.SessionHandler{
		Authenticator:  authenticator,
		ProfileService: ctn.Get("profile-service").(profile.Service),
		SessionService: ctn.Get("session-service").(session.Service),
	}
	handlers := []server.Handler{
		metricHandler,
		healthHandler,
		authenticationHandler,
		profileHandler,
		sessionHandler,
	}

	server := server.NewServer(handlers)
	srv := server.CreateHTTPServer()

	// Handle SIGINT, SIGTERN, SIGHUP signal from OS
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-termChan
		log.Warn("Receiving signal, Shutting down server")
		srv.Close()
	}()

	log.WithField("Port", server.Port).Info("Userland API Server is running")
	log.Fatal(srv.ListenAndServe())
}
