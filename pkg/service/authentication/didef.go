package authentication

import (
	"github.com/AdhityaRamadhanus/userland"
	mailing "github.com/AdhityaRamadhanus/userland/pkg/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/sarulabs/di"
)

var (
	ServiceBuilder = di.Def{
		Name:  "authentication-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			userRepository := ctn.Get("user-repository").(userland.UserRepository)
			keyValueService := ctn.Get("keyvalue-service").(userland.KeyValueService)
			mailingClient := ctn.Get("mailing-client").(mailing.Client)
			return service{
				userRepository:  userRepository,
				keyValueService: keyValueService,
				mailingClient:   mailingClient,
			}, nil
		},
	}

	ServiceInstrumentorBuilder = di.Def{
		Name:  "authentication-instrumentor-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			authenticationService := ctn.Get("authentication-service").(Service)
			return NewInstrumentorService(
				metrics.PrometheusRequestCounter("api", "authentication_service", MetricKeys),
				metrics.PrometheusRequestLatency("api", "authentication_service", MetricKeys),
				authenticationService,
			), nil
		},
	}
)
