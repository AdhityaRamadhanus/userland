package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	server "github.com/AdhityaRamadhanus/userland/pkg/server/api"
	"github.com/AdhityaRamadhanus/userland/pkg/server/api/handlers"
	"github.com/AdhityaRamadhanus/userland/pkg/service/authentication"
	"github.com/AdhityaRamadhanus/userland/pkg/service/event"
	"github.com/AdhityaRamadhanus/userland/pkg/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/service/session"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/gcs"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func buildConfig() *config.Configuration {
	envPath := ".env"
	if err := godotenv.Load(envPath); err != nil {
		logrus.Fatalf("godotenv.Load(%q) err = %v", envPath, err)
	}

	yamlPath := "config.yaml"
	envPrefix := ""
	c, err := config.Build(yamlPath, envPrefix)
	if err != nil {
		logrus.Fatalf("config.Build(%q, %q) err = %v", yamlPath, envPrefix, err)
	}

	return c
}

func setupLogger(cfg config.LogConfig) {
	switch cfg.Format {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	switch cfg.Level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	cfg := buildConfig()
	setupLogger(cfg.Log)

	logrus.Debug("Connecting to postgres at", cfg.Postgres)
	pgConn, err := postgres.CreateConnection(cfg.Postgres)
	if err != nil {
		logrus.Fatalf("postgres.CreateConnection() err = %v", err)
	}
	logrus.Debug("Connecting to redis at", cfg.Redis)
	redisClient, err := redis.CreateClient(cfg.Redis, 0)
	if err != nil {
		logrus.Fatalf("redis.CreateClient(cfg, 0) err = %v", err)
	}
	redisRateClient, err := redis.CreateClient(cfg.Redis, 1)
	if err != nil {
		logrus.Fatalf("redis.CreateClient(cfg, 1) err = %v", err)
	}
	ctx := context.Background()
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		logrus.Fatalf("(GCS) storage.NewClient(ctx) err = %v", err)
	}

	// mailing Client
	mailClient := mailing.NewMailingClient(cfg.Mail.Host, mailing.WithClientTimeout(5*time.Second), mailing.WithBasicAuth(cfg.Mail.AuthUser, cfg.Mail.AuthPassword))
	// repositories
	userRepository := postgres.NewUserRepository(pgConn)
	eventRepository := postgres.NewEventRepository(pgConn)
	sessionRepository := redis.NewSessionRepository(redisClient)
	keyValueSvc := redis.NewKeyValueService(redisClient)
	objectStorageSvc := gcs.NewObjectStorageService(gcsClient, cfg.GCP.BucketName)

	// services
	authSvc := authentication.NewService(
		authentication.WithKeyValueService(keyValueSvc),
		authentication.WithMailingClient(mailClient),
		authentication.WithUserRepository(userRepository),
	)
	// authInstSvc := authentication.NewInstrumentorService(metrics.PrometheusRequestLatency("service", "authentication", authentication.MetricKeys), authSvc)

	profileSvc := profile.NewService(
		profile.WithKeyValueService(keyValueSvc),
		profile.WithMailingClient(mailClient),
		profile.WithObjectStorageService(objectStorageSvc),
		profile.WithUserRepository(userRepository),
	)

	sessionSvc := session.NewService(session.WithKeyValueService(keyValueSvc), session.WithSessionRepository(sessionRepository))
	eventSvc := event.NewService(event.WithEventRepository(eventRepository))

	authenticator := middlewares.TokenAuth(keyValueSvc)
	ratelimiter := middlewares.RateLimit(redisRateClient)

	healthHandler := handlers.HealthzHandler{}
	metricHandler := handlers.MetricHandler{}
	authenticationHandler := handlers.AuthenticationHandler{
		RateLimiter:           ratelimiter,
		Authenticator:         authenticator,
		Authorization:         middlewares.Authorize,
		ProfileService:        profileSvc,
		AuthenticationService: authSvc,
		SessionService:        sessionSvc,
		EventService:          eventSvc,
	}
	profileHandler := handlers.ProfileHandler{
		Authorization:  middlewares.Authorize,
		RateLimiter:    ratelimiter,
		Authenticator:  authenticator,
		ProfileService: profileSvc,
		EventService:   eventSvc,
	}
	sessionHandler := handlers.SessionHandler{
		Authorization:  middlewares.Authorize,
		Authenticator:  authenticator,
		ProfileService: profileSvc,
		SessionService: sessionSvc,
	}

	server := server.NewServer(cfg.API, metricHandler, healthHandler, authenticationHandler, profileHandler, sessionHandler)
	srv := server.CreateHTTPServer()

	// Handle SIGINT, SIGTERN, SIGHUP signal from OS
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-termChan
		logrus.Warn("Receiving signal, Shutting down server")
		srv.Close()
	}()

	logrus.WithField("Address", server.Address).Info("Userland API Server is running")
	logrus.Fatal(srv.ListenAndServe())
}
