package mailing

import (
	"github.com/AdhityaRamadhanus/userland/pkg/service/mailing"
	"github.com/stretchr/testify/mock"
)

type MailingService struct {
	mock.Mock
}

func (m MailingService) SendOTPEmail(recipient mailing.MailAddress, otpType string, otp string) error {
	args := m.Called(recipient, otpType, otp)

	return args.Get(0).(error)
}

func (m MailingService) SendVerificationEmail(recipient mailing.MailAddress, verificationLink string) error {
	args := m.Called(recipient, verificationLink)

	return args.Get(0).(error)
}
