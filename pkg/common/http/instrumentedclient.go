package http

import (
	"net/http"
	"strconv"
	"time"

	_metrics "github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/go-kit/kit/metrics"
)

type InstrumentedClient struct {
	client         *http.Client
	requestLatency metrics.Histogram
}

func WithClientTimeout(timeout time.Duration) func(ic *InstrumentedClient) {
	return func(ic *InstrumentedClient) {
		ic.client.Timeout = timeout
	}
}

func NewInstrumentedClient(namespace string, options ...func(*InstrumentedClient)) *InstrumentedClient {
	client := &InstrumentedClient{
		client:         &http.Client{},
		requestLatency: _metrics.PrometheusRequestLatency("http_client", namespace, []string{"method", "route", "status_code"}),
	}
	for _, option := range options {
		option(client)
	}

	return client
}

func (i InstrumentedClient) Do(req *http.Request) (res *http.Response, err error) {
	defer func(begin time.Time) {
		if res != nil {
			i.requestLatency.
				With("method", req.Method, "route", req.URL.Path, "status_code", strconv.Itoa(res.StatusCode)).
				Observe(time.Since(begin).Seconds())
		}
	}(time.Now())

	return i.client.Do(req)
}
