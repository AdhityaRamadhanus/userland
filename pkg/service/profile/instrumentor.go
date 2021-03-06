package profile

import (
	"io"
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

func (s instrumentorService) ProfileByEmail(email string) (userland.User, error) {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "ProfileByEmail").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ProfileByEmail(email)
}

func (s instrumentorService) Profile(userID int) (userland.User, error) {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "Profile").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Profile(userID)
}

func (s instrumentorService) SetProfile(user userland.User) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "SetProfile").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.SetProfile(user)
}

func (s instrumentorService) RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error) {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "RequestChangeEmail").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.RequestChangeEmail(user, newEmail)
}

func (s instrumentorService) ChangeEmail(user userland.User, verificationID string) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "ChangeEmail").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ChangeEmail(user, verificationID)
}

func (s instrumentorService) ChangePassword(user userland.User, oldPassword string, newPassword string) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "ChangePassword").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ChangePassword(user, oldPassword, newPassword)
}

func (s instrumentorService) EnrollTFA(user userland.User) (secret string, qrcodeImageBase64 string, err error) {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "EnrollTFA").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.EnrollTFA(user)
}

func (s instrumentorService) ActivateTFA(user userland.User, secret string, code string) ([]string, error) {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "ActivateTFA").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.ActivateTFA(user, secret, code)
}

func (s instrumentorService) RemoveTFA(user userland.User, currPassword string) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "RemoveTFA").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.RemoveTFA(user, currPassword)
}

func (s instrumentorService) DeleteAccount(user userland.User, currPassword string) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "DeleteAccount").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.DeleteAccount(user, currPassword)
}

func (s instrumentorService) SetProfilePicture(user userland.User, image io.Reader) error {
	defer func(begin time.Time) {
		s.requestLatency.With("method", "SetProfilePicture").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.SetProfilePicture(user, image)
}
