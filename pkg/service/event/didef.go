package event

import (
	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/sarulabs/di"
)

var (
	ServiceBuilder = di.Def{
		Name:  "event-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			eventRepository := ctn.Get("event-repository").(userland.EventRepository)

			return service{
				eventRepository: eventRepository,
			}, nil
		},
	}

	ServiceInstrumentorBuilder = di.Def{
		Name:  "event-instrumentor-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			eventService := ctn.Get("event-service").(Service)
			return NewInstrumentorService(
				metrics.PrometheusRequestCounter("api", "event_service", MetricKeys),
				metrics.PrometheusRequestLatency("api", "event_service", MetricKeys),
				eventService,
			), nil
		},
	}
)
