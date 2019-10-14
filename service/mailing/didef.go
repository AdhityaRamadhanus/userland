package mailing

import (
	"github.com/AdhityaRamadhanus/userland/metrics"
	"github.com/gocraft/work"
	"github.com/sarulabs/di"

	mailjet "github.com/mailjet/mailjet-apiv3-go"
)

var (
	WorkerBuilder = di.Def{
		Name:  "mailing-worker",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			mailjetClient := ctn.Get("mailjet-client").(*mailjet.Client)
			return NewWorker(mailjetClient), nil
		},
	}

	ServiceBuilder = di.Def{
		Name:  "mailing-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			workEnqueuer := ctn.Get("work-enqueuer").(*work.Enqueuer)

			return service{
				producer: workEnqueuer,
			}, nil
		},
	}

	ServiceInstrumentorBuilder = di.Def{
		Name:  "mailing-instrumentor-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			mailingService := ctn.Get("mailing-service").(Service)
			return NewInstrumentorService(
				metrics.PrometheusRequestCounter("api", "mailing_service", MetricKeys),
				metrics.PrometheusRequestLatency("api", "mailing_service", MetricKeys),
				mailingService,
			), nil
		},
	}
)
