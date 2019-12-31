package mailing

import (
	"encoding/json"
	"fmt"

	"github.com/gocraft/work"
	"github.com/pkg/errors"
)

type MailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type SendEmailOption struct {
	From         MailAddress            `json:"from"`
	To           []MailAddress          `json:"to"`
	Subject      string                 `json:"subject"`
	Template     string                 `json:"template"`
	TemplateArgs map[string]interface{} `json:"args"`
}

type service struct {
	queueName   string
	emailSender string
	producer    *work.Enqueuer
}

type Service interface {
	SendOTPEmail(recipient MailAddress, otpType string, otp string) error
	SendVerificationEmail(recipient MailAddress, verificationLink string) error
}

func NewService(queueName, emailSender string, enqueuer *work.Enqueuer) Service {
	return &service{
		queueName:   queueName,
		emailSender: emailSender,
		producer:    enqueuer,
	}
}

func (s service) SendOTPEmail(recipient MailAddress, otpType string, otp string) error {
	opts := SendEmailOption{
		From: MailAddress{
			Name:    "OTP Email from Userland",
			Address: s.emailSender,
		},
		To: []MailAddress{
			{
				Name:    recipient.Name,
				Address: recipient.Address,
			},
		},
		Subject:  fmt.Sprintf("OTP for %s", otpType),
		Template: "otp",
		TemplateArgs: map[string]interface{}{
			"otp":       otp,
			"otp_type":  otpType,
			"recipient": recipient.Name,
		},
	}

	// convert struct to json
	work := work.Q{}
	jsonBytes, err := json.Marshal(opts)
	if err != nil {
		return errors.Wrap(err, "json.Marshal() err")
	}
	json.Unmarshal(jsonBytes, &work)

	if _, err := s.producer.Enqueue(s.queueName, work); err != nil {
		return errors.Wrap(err, "producer.Enqueue() err")
	}

	return nil
}

func (s service) SendVerificationEmail(recipient MailAddress, verificationLink string) (err error) {
	opts := SendEmailOption{
		From: MailAddress{
			Name:    "Verification Email from Userland",
			Address: s.emailSender,
		},
		To: []MailAddress{
			{
				Name:    recipient.Name,
				Address: recipient.Address,
			},
		},
		Subject:  fmt.Sprintf("Verification for %s", recipient.Address),
		Template: "email_verification",
		TemplateArgs: map[string]interface{}{
			"verification_link": verificationLink,
			"recipient":         recipient.Name,
		},
	}

	// convert struct to json
	// convert struct to json
	work := work.Q{}
	jsonBytes, err := json.Marshal(opts)
	if err != nil {
		return errors.Wrap(err, "json.Marshal() err")
	}
	json.Unmarshal(jsonBytes, &work)

	if _, err := s.producer.Enqueue(s.queueName, work); err != nil {
		return errors.Wrap(err, "producer.Enqueue() err")
	}

	return nil
}
