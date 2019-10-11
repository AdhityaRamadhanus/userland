package authentication

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

func (s instrumentorService) Register(user userland.User) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Register").Add(1)
		s.requestLatency.With("method", "Register").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Register(user)
}

func (s instrumentorService) RequestVerification(verificationType string, email string) (verificationID string, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "RequestVerification").Add(1)
		s.requestLatency.With("method", "RequestVerification").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.RequestVerification(verificationType, email)
}

func (s instrumentorService) VerifyAccount(verificationID string, email string, code string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "VerifyAccount").Add(1)
		s.requestLatency.With("method", "VerifyAccount").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.VerifyAccount(verificationID, email, code)
}

func (s instrumentorService) Login(email, password string) (requireTFA bool, accessToken security.AccessToken, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "Login").Add(1)
		s.requestLatency.With("method", "Login").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Login(email, password)
}

func (s instrumentorService) VerifyTFA(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "VerifyTFA").Add(1)
		s.requestLatency.With("method", "VerifyTFA").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.VerifyTFA(tfaToken, userID, code)
}

func (s instrumentorService) VerifyTFABypass(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "VerifyTFABypass").Add(1)
		s.requestLatency.With("method", "VerifyTFABypass").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.VerifyTFABypass(tfaToken, userID, code)
}

func (s instrumentorService) ForgotPassword(email string) (verificationID string, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "ForgotPassword").Add(1)
		s.requestLatency.With("method", "ForgotPassword").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ForgotPassword(email)
}

func (s instrumentorService) ResetPassword(forgotPassToken string, newPassword string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "ResetPassword").Add(1)
		s.requestLatency.With("method", "ResetPassword").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ResetPassword(forgotPassToken, newPassword)
}
