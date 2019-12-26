package mailing

import (
	"github.com/AdhityaRamadhanus/userland/pkg/service/mailing"
)

type SimpleMailingService struct {
	CalledMethods map[string]bool
}

func (m SimpleMailingService) SendOTPEmail(recipient mailing.MailAddress, otpType string, otp string) error {
	m.CalledMethods["SendOTPEmail"] = true
	return nil
}

func (m SimpleMailingService) SendVerificationEmail(recipient mailing.MailAddress, verificationLink string) error {
	m.CalledMethods["SendVerificationEmail"] = true
	return nil
}
