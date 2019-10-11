package session

import (
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/security"
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

func (s instrumentorService) CreateSession(userID int, session userland.Session) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "CreateSession").Add(1)
		s.requestLatency.With("method", "CreateSession").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.CreateSession(userID, session)
}

func (s instrumentorService) ListSession(userID int) (userland.Sessions, error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "ListSession").Add(1)
		s.requestLatency.With("method", "ListSession").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ListSession(userID)
}

func (s instrumentorService) EndSession(userID int, currentSessionID string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "EndSession").Add(1)
		s.requestLatency.With("method", "EndSession").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.EndSession(userID, currentSessionID)
}

func (s instrumentorService) EndOtherSessions(userID int, currentSessionID string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "EndOtherSessions").Add(1)
		s.requestLatency.With("method", "EndOtherSessions").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.EndOtherSessions(userID, currentSessionID)
}

func (s instrumentorService) CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "CreateRefreshToken").Add(1)
		s.requestLatency.With("method", "CreateRefreshToken").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.CreateRefreshToken(user, currentSessionID)
}

func (s instrumentorService) CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "CreateNewAccessToken").Add(1)
		s.requestLatency.With("method", "CreateNewAccessToken").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.CreateNewAccessToken(user, refreshTokenID)
}
