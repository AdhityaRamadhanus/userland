//+build unit

package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_http "github.com/AdhityaRamadhanus/userland/pkg/common/http"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/server/mailing/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestMailingHandler(t *testing.T) {
	mailingService := mailing.SimpleMailingService{CalledMethods: map[string]bool{}}

	mailingHandler := handlers.MailingHandler{
		Authenticator:  middlewares.Bypass,
		MailingService: mailingService,
	}
	router := mux.NewRouter().StrictSlash(true)
	mailingHandler.RegisterRoutes(router)

	ts := httptest.NewServer(middlewares.ClientParser(router))
	defer ts.Close()

	testCases := []struct {
		Name           string
		Method         string
		Path           string
		RequestBody    map[string]interface{}
		ExpectedStatus int
	}{
		{
			Name:   "Test sendEmailOTP",
			Method: http.MethodPost,
			Path:   "api/mail/otp",
			RequestBody: map[string]interface{}{
				"recipient_name": "Adhitya Ramadhanus",
				"recipient":      "adhitya.ramadhanus@gmail.com",
				"type":           "TFA Verification",
				"otp":            "123123",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test sendEmailVerification",
			Method: http.MethodPost,
			Path:   "api/mail/verification",
			RequestBody: map[string]interface{}{
				"recipient_name":    "Adhitya Ramadhanus",
				"recipient":         "adhitya.ramadhanus@gmail.com",
				"verification_link": "testaja",
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", ts.URL, testCase.Path)
			req, _ := _http.CreateJSONRequest(testCase.Method, url, testCase.RequestBody)
			res, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			assert.Equal(t, res.StatusCode, testCase.ExpectedStatus)
		})
	}
}
