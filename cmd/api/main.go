package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_gcs "cloud.google.com/go/storage"
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
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

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
	pgConnString := postgres.CreateConnectionString()
	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := _redis.NewClient(&_redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,                           // use default DB
	})
	redisRateLimitClient := _redis.NewClient(&_redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       2,                           // use default DB
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Fatal("Failed to connect to redis")
	}

	ctx := context.Background()
	gcsClient, err := _gcs.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mailingClient := mailing.NewMailingClient(
		os.Getenv("USERLAND_MAIL_HOST"),
		mailing.WithClientTimeout(time.Second*5),
		mailing.WithBasicAuth(os.Getenv("MAIL_SERVICE_BASIC_USER"), os.Getenv("MAIL_SERVICE_BASIC_PASS")),
	)

	// Repositories
	keyValueService := redis.NewKeyValueService(redisClient)
	userRepository := postgres.NewUserRepository(db)
	eventRepository := postgres.NewEventRepository(db)
	sessionRepository := redis.NewSessionRepository(redisClient)
	objectStorageService := gcs.NewObjectStorageService(gcsClient, "userland_cdn")

	eventService := event.NewService(
		event.WithEventRepository(eventRepository),
	)
	authenticationService := authentication.NewService(
		authentication.WithUserRepository(userRepository),
		authentication.WithKeyValueService(keyValueService),
		authentication.WithMailingClient(mailingClient),
	)
	profileService := profile.NewService(
		profile.WithEventRepository(eventRepository),
		profile.WithUserRepository(userRepository),
		profile.WithKeyValueService(keyValueService),
		profile.WithObjectStorageService(objectStorageService),
	)
	sessionService := session.NewService(keyValueService, sessionRepository)

	authenticator := middlewares.NewAuthenticator(keyValueService)

	ratelimiter := middlewares.NewRateLimiter(redisRateLimitClient)
	healthHandler := handlers.HealthzHandler{}
	authenticationHandler := handlers.AuthenticationHandler{
		RateLimiter:           ratelimiter,
		Authenticator:         authenticator,
		ProfileService:        profileService,
		AuthenticationService: authenticationService,
		SessionService:        sessionService,
		EventService:          eventService,
	}
	profileHandler := handlers.ProfileHandler{
		RateLimiter:    ratelimiter,
		Authenticator:  authenticator,
		ProfileService: profileService,
		EventService:   eventService,
	}
	sessionHandler := handlers.SessionHandler{
		Authenticator:  authenticator,
		ProfileService: profileService,
		SessionService: sessionService,
	}
	handlers := []server.Handler{
		healthHandler,
		authenticationHandler,
		profileHandler,
		sessionHandler,
	}

	server := server.NewServer(handlers)
	srv := server.CreateHttpServer()

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
