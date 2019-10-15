package mailing

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"

	"github.com/go-errors/errors"

	"fmt"
	"net/http"
	"time"
)

type Client interface {
	SendOTPEmail(recipientAddress string, recipientName string, otpType string, otp string) error
	SendVerificationEmail(recipientAddress string, recipientName string, verificationLink string) error
}

type client struct {
	username   string
	password   string
	baseURL    string
	httpClient *http.Client
}

func WithBasicAuth(username, password string) func(client *client) {
	return func(client *client) {
		client.username = username
		client.password = password
	}
}

func WithClientTimeout(timeout time.Duration) func(client *client) {
	return func(client *client) {
		client.httpClient.Timeout = timeout
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
	jsonBytes, _ := json.Marshal(requestBody)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password))))

	client := &http.Client{}
	resp, err := client.Do(req)
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
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password))))

	client := &http.Client{}
	resp, err := client.Do(req)
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
