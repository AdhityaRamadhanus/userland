package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	server "github.com/AdhityaRamadhanus/userland/pkg/server/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/server/mailing/handlers"
	"github.com/AdhityaRamadhanus/userland/pkg/service/mailing"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/prometheus/common/log"
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

	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				redisAddr,
				redis.DialDatabase(1),
			)
		},
	}
	enqueuer := work.NewEnqueuer(cfg.Mail.WorkerSpace, redisPool)

	mailjetClient := mailjet.NewMailjetClient(cfg.Mailjet.PublicKey, cfg.Mailjet.PrivateKey)
	mailingWorker := mailing.NewWorker(mailjetClient)
	mailingService := mailing.NewService(enqueuer)

	basicAuthenticator := middlewares.BasicAuth(cfg.Mail.AuthUser, cfg.Mail.AuthPassword)
	healthHandler := handlers.HealthzHandler{}
	mailingHandler := handlers.MailingHandler{
		Authenticator:  basicAuthenticator,
		MailingService: mailingService,
	}

	server := server.NewServer(cfg.Mail, healthHandler, mailingHandler)
	srv := server.CreateHTTPServer()

	pool := work.NewWorkerPool(struct{}{}, uint(runtime.NumCPU()), cfg.Mail.WorkerSpace, redisPool)
	// Map the name of jobs to handler functions
	pool.Job(cfg.Mail.Queue, mailingWorker.EnquiryJob)

	// Handle SIGINT, SIGTERN, SIGHUP signal from OS
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-termChan
		log.Warn("Receiving signal, Shutting down server")
		srv.Close()
	}()

	logrus.Info("Start processing jobs")
	pool.Start()
	logrus.WithField("Address", server.Address).Info("Userland Mail API Server is running")
	logrus.Fatal(srv.ListenAndServe())
}
