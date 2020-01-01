package mailing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	_http "github.com/AdhityaRamadhanus/userland/pkg/common/http"
	"github.com/pkg/errors"
)

var (
	ErrSendEmailFailed = errors.New("Failed to send email")
)

type Client interface {
	SendOTPEmail(recipientAddress string, recipientName string, otpType string, otp string) error
	SendVerificationEmail(recipientAddress string, recipientName string, verificationLink string) error
}

type client struct {
	username   string
	password   string
	baseURL    string
	httpClient _http.Client
}

func WithBasicAuth(username, password string) func(client *client) {
	return func(client *client) {
		client.username = username
		client.password = password
	}
}

func WithHTTPClient(c _http.Client) func(client *client) {
	return func(client *client) {
		client.httpClient = c
	}
}

func NewMailingClient(baseURL string, options ...func(*client)) Client {
	client := &client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
	for _, option := range options {
		option(client)
	}

	return client
}

func (c client) SendOTPEmail(recipientAddress string, recipientName string, otpType string, otp string) error {
	url := fmt.Sprintf("%s/api/mail/otp", c.baseURL)

	requestBody := map[string]interface{}{
		"recipient":      recipientAddress,
		"recipient_name": recipientName,
		"otp":            otp,
		"type":           otpType,
	}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return errors.Wrapf(err, "json.Marshal() err")
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.Wrapf(err, "http.NewRequest() err")
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "httpClient.Do() err")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Wrapf(ErrSendEmailFailed, "%s return status code = %d", url, resp.StatusCode)
	}

	return nil
}

func (c client) SendVerificationEmail(recipientAddress string, recipientName string, verificationLink string) error {
	url := fmt.Sprintf("%s/api/mail/verification", c.baseURL)

	requestBody := map[string]interface{}{
		"recipient":         recipientAddress,
		"recipient_name":    recipientName,
		"verification_link": verificationLink,
	}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return errors.Wrapf(err, "json.Marshal() err")
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.Wrapf(err, "http.NewRequest() err")
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "httpClient.Do() err")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Wrapf(ErrSendEmailFailed, "%s return status code = %d", url, resp.StatusCode)
	}

	return nil
}
