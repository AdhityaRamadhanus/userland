package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdhityaRamadhanus/userland/authentication"
	"github.com/AdhityaRamadhanus/userland/profile"
	"github.com/AdhityaRamadhanus/userland/server"
	"github.com/AdhityaRamadhanus/userland/server/handlers"
	"github.com/AdhityaRamadhanus/userland/server/middlewares"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	_redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func main() {
	godotenv.Load()
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

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.WithError(err).Error("Failed to connect to redis")
	}

	// Repositories
	keyValueService := redis.NewKeyValueService(redisClient)
	userRepository := postgres.NewUserRepository(db)
	eventRepository := postgres.NewEventRepository(db)

	authenticationService := authentication.NewService(userRepository, keyValueService)
	profileService := profile.NewService(eventRepository, userRepository, keyValueService)

	authenticator := middlewares.NewAuthenticator(keyValueService)
	healthHandler := handlers.HealthzHandler{}
	authenticationHandler := handlers.AuthenticationHandler{
		Authenticator:         authenticator,
		AuthenticationService: authenticationService,
	}
	profileHandler := handlers.ProfileHandler{
		Authenticator:  authenticator,
		ProfileService: profileService,
	}
	handlers := []server.Handler{
		healthHandler,
		authenticationHandler,
		profileHandler,
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
