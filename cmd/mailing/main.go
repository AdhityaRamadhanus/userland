package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	server "github.com/AdhityaRamadhanus/userland/server/mailing"
	"github.com/AdhityaRamadhanus/userland/server/mailing/handlers"
	"github.com/AdhityaRamadhanus/userland/service/mailing"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	log "github.com/sirupsen/logrus"
)

func main() {
	godotenv.Load()
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

	workerNamespace := "userland-mail-worker"
	enqueuer := work.NewEnqueuer(workerNamespace, redisPool)
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_APIKEY_PUBLIC"), os.Getenv("MAILJET_APIKEY_PRIVATE"))

	mailingService := mailing.NewService(enqueuer)
	mailingWorker := mailing.NewWorker(mailjetClient)

	healthHandler := handlers.HealthzHandler{}
	mailingHandler := handlers.MailHandler{
		MailingService: mailingService,
	}
	handlers := []server.Handler{
		healthHandler,
		mailingHandler,
	}

	server := server.NewServer(handlers)
	srv := server.CreateHttpServer()

	pool := work.NewWorkerPool(struct{}{}, uint(runtime.NumCPU()), workerNamespace, redisPool)
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