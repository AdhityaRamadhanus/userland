package session

import (
	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/metrics"
	"github.com/sarulabs/di"
)

var (
	ServiceBuilder = di.Def{
		Name:  "session-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			keyValueService := ctn.Get("keyvalue-service").(userland.KeyValueService)
			sessionRepository := ctn.Get("session-repository").(userland.SessionRepository)

			return service{
				keyValueService:   keyValueService,
				sessionRepository: sessionRepository,
			}, nil
		},
	}

	ServiceInstrumentorBuilder = di.Def{
		Name:  "session-instrumentor-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			sessionService := ctn.Get("session-service").(Service)
			return NewInstrumentorService(
				metrics.PrometheusRequestCounter("api", "session_service", MetricKeys),
				metrics.PrometheusRequestLatency("api", "session_service", MetricKeys),
				sessionService,
			), nil
		},
	}
)
