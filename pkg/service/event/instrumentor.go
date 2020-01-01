package event

import (
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/go-kit/kit/metrics"
)

var (
	MetricKeys = []string{"method"}
)

type instrumentorService struct {
	requestLatency metrics.Histogram
	next           Service
}

//Service provide an interface to story domain service

func NewInstrumentorService(latency metrics.Histogram, s Service) Service {
	service := &instrumentorService{
		requestLatency: latency,
		next:           s,
	}

	return service
}

func (s *instrumentorService) Log(eventName string, userID int, clientInfo map[string]interface{}) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "Log").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Log(eventName, userID, clientInfo)
}

func (s *instrumentorService) ListEvents(filter userland.EventFilterOptions, paging userland.EventPagingOptions) (events userland.Events, count int, err error) {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "ListEvents").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ListEvents(filter, paging)
}

func (s *instrumentorService) DeleteEventsByUserID(userID int) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "DeleteEventsByUserID").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.DeleteEventsByUserID(userID)
}
