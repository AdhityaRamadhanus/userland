package mailing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	_http "github.com/AdhityaRamadhanus/userland/pkg/common/http"
	"github.com/pkg/errors"
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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to send OTP Email")
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
	jsonBytes, _ := json.Marshal(requestBody)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var decodedResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&decodedResponse)
		return errors.New(fmt.Sprintf("Failed to send Verification Email reponse code %d, paload %v", resp.StatusCode, decodedResponse))
	}

	return nil
}
