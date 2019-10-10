package mailing

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gocraft/work"
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
	producer *work.Enqueuer
}

type Service interface {
	SendOTPEmail(recipient MailAddress, otpType string, otp string) error
}

func NewService(enqueuer *work.Enqueuer) Service {
	return &service{
		producer: enqueuer,
	}
}

func (s service) SendOTPEmail(recipient MailAddress, otpType string, otp string) error {
	queueName := os.Getenv("EMAIL_QUEUE")
	opts := SendEmailOption{
		From: MailAddress{
			Name:    "OTP Email from Userland",
			Address: "adhitya.ramadhanus@gmail.com",
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
	jsonBytes, _ := json.Marshal(opts)
	json.Unmarshal(jsonBytes, &work)

	_, err := s.producer.Enqueue(queueName, work)
	return err
}
