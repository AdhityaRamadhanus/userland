package event

import (
	"time"

	"github.com/go-kit/kit/metrics"
)

var (
	MetricKeys = []string{"method"}
)

type instrumentorService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

//Service provide an interface to story domain service

func NewInstrumentorService(counter metrics.Counter, latency metrics.Histogram, s Service) Service {
	service := &instrumentorService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}

	return service
}

func (s *instrumentorService) Log(eventName string, userID int, clientInfo map[string]interface{}) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Log").Add(1)
		s.requestLatency.With("method", "Log").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Log(eventName, userID, clientInfo)
}
