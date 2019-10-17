package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	server "github.com/AdhityaRamadhanus/userland/server/mailing"
	"github.com/AdhityaRamadhanus/userland/server/mailing/handlers"
	"github.com/AdhityaRamadhanus/userland/service/mailing"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/sarulabs/di"
	log "github.com/sirupsen/logrus"
)

func buildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		di.Def{
			Name:  "redis-pool-connection",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				redisAddr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
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
				return redisPool, nil
			},
		},
		di.Def{
			Name:  "work-enqueuer",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				redisPool := ctn.Get("redis-pool-connection").(*redis.Pool)
				enqueuer := work.NewEnqueuer(os.Getenv("MAIL_WORKER_SPACE"), redisPool)
				return enqueuer, nil
			},
		},
		di.Def{
			Name:  "mailjet-client",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_APIKEY_PUBLIC"), os.Getenv("MAILJET_APIKEY_PRIVATE"))
				return mailjetClient, nil
			},
		},
		mailing.ServiceBuilder,
		mailing.ServiceInstrumentorBuilder,
		mailing.WorkerBuilder,
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

	redisPool := ctn.Get("redis-pool-connection").(*redis.Pool)
	mailingWorker := ctn.Get("mailing-worker").(*mailing.Worker)

	basicAuthenticator := middlewares.BasicAuth(os.Getenv("MAIL_SERVICE_BASIC_USER"), os.Getenv("MAIL_SERVICE_BASIC_PASS"))
	healthHandler := handlers.HealthzHandler{}
	mailingHandler := handlers.MailingHandler{
		Authenticator:  basicAuthenticator,
		MailingService: ctn.Get("mailing-service").(mailing.Service),
	}
	handlers := []server.Handler{
		healthHandler,
		mailingHandler,
	}

	server := server.NewServer(handlers)
	srv := server.CreateHTTPServer()

	pool := work.NewWorkerPool(struct{}{}, uint(runtime.NumCPU()), os.Getenv("MAIL_WORKER_SPACE"), redisPool)
	// Map the name of jobs to handler functions
	pool.Job(os.Getenv("EMAIL_QUEUE"), mailingWorker.EnquiryJob)

	// Handle SIGINT, SIGTERN, SIGHUP signal from OS
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-termChan
		log.Warn("Receiving signal, Shutting down server")
		srv.Close()
	}()

	log.Info("Start processing jobs")
	pool.Start()
	log.WithField("Port", server.Port).Info("Userland Mail API Server is running")
	log.Fatal(srv.ListenAndServe())
}
