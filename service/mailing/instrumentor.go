package mailing

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

func (s instrumentorService) SendOTPEmail(recipient MailAddress, otpType string, otp string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "SendOTPEmail").Add(1)
		s.requestLatency.With("method", "SendOTPEmail").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.SendOTPEmail(recipient, otpType, otp)
}

func (s instrumentorService) SendVerificationEmail(recipient MailAddress, verificationLink string) error {
	defer func(begin time.Time) {
		s.requestCount.With("method", "SendVerificationEmail").Add(1)
		s.requestLatency.With("method", "SendVerificationEmail").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.SendVerificationEmail(recipient, verificationLink)
}
