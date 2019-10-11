package metrics

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// return counter and histogram
func PrometheusRequestCounter(namespace, service string, fieldKeys []string) metrics.Counter {
	return kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: service,
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
}

func PrometheusRequestLatency(namespace, service string, fieldKeys []string) metrics.Histogram {
	return kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: service,
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

}
