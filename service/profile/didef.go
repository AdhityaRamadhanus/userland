package profile

import (
	"github.com/AdhityaRamadhanus/userland"
	mailing "github.com/AdhityaRamadhanus/userland/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/common/metrics"
	"github.com/sarulabs/di"
)

var (
	ServiceBuilder = di.Def{
		Name:  "profile-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			userRepository := ctn.Get("user-repository").(userland.UserRepository)
			keyValueService := ctn.Get("keyvalue-service").(userland.KeyValueService)
			mailingClient := ctn.Get("mailing-client").(mailing.Client)
			objectStorageService := ctn.Get("objectstorage-service").(userland.ObjectStorageService)
			eventRepository := ctn.Get("event-repository").(userland.EventRepository)

			return service{
				userRepository:       userRepository,
				keyValueService:      keyValueService,
				mailingClient:        mailingClient,
				eventRepository:      eventRepository,
				objectStorageService: objectStorageService,
			}, nil
		},
	}

	ServiceInstrumentorBuilder = di.Def{
		Name:  "profile-instrumentor-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			profileService := ctn.Get("profile-service").(Service)
			return NewInstrumentorService(
				metrics.PrometheusRequestCounter("api", "profile_service", MetricKeys),
				metrics.PrometheusRequestLatency("api", "profile_service", MetricKeys),
				profileService,
			), nil
		},
	}
)
