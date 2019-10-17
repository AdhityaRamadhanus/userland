//+build unit

package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/mocks/middlewares"
	"github.com/AdhityaRamadhanus/userland/mocks/service/authentication"
	"github.com/AdhityaRamadhanus/userland/mocks/service/event"
	"github.com/AdhityaRamadhanus/userland/mocks/service/profile"
	"github.com/AdhityaRamadhanus/userland/mocks/service/session"
	"github.com/AdhityaRamadhanus/userland/server/api/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func createHttpJSONRequest(method string, path string, requestBody interface{}) (*http.Request, error) {
	var httpReq *http.Request
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		jsonReqBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, errors.New("Failed to marshal request body")
		}
		httpReq, err = http.NewRequest(method, path, bytes.NewBuffer(jsonReqBody))
		if err != nil {
			return nil, errors.New("Failed to create request body")
		}
		httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	case http.MethodGet:
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			return nil, errors.New("Failed to marshal get request body")
		}
		httpReq = req
	}

	return httpReq, nil
}

func TestAuthenticationHandler(t *testing.T) {
	profileService := profile.SimpleProfileService{CalledMethods: map[string]bool{}}
	sessionService := session.SimpleSessionService{CalledMethods: map[string]bool{}}
	authenticationService := authentication.SimpleAuthenticationService{CalledMethods: map[string]bool{}}
	eventService := event.SimpleEventService{CalledMethods: map[string]bool{}}

	authenticationHandler := handlers.AuthenticationHandler{
		RateLimiter:           middlewares.BypassWithArgs,
		Authenticator:         middlewares.AuthenticationWithCustomScope(security.TFATokenScope),
		ProfileService:        profileService,
		AuthenticationService: authenticationService,
		SessionService:        sessionService,
		EventService:          eventService,
	}
	router := mux.NewRouter().StrictSlash(true)
	authenticationHandler.RegisterRoutes(router)

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
			Name:   "Test registerUser",
			Method: http.MethodPost,
			Path:   "api/auth/register",
			RequestBody: map[string]interface{}{
				"email":              "adhitya.ramadhanus@gmail.com",
				"password":           "test123",
				"password_confirmed": "test123",
				"fullname":           "adhitya ramadhanus",
			},
			ExpectedStatus: http.StatusCreated,
		},
		{
			Name:   "Test requestVerification",
			Method: http.MethodPost,
			Path:   "api/auth/verification",
			RequestBody: map[string]interface{}{
				"type":      "email.verify",
				"recipient": "adhitya.ramadhanus@gmail.com",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test verifyAccount",
			Method: http.MethodPatch,
			Path:   "api/auth/verification",
			RequestBody: map[string]interface{}{
				"verification_id": "asdasdasdasdasdasd",
				"code":            "test123",
				"email":           "adhitya.ramadhanus@gmail.com",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test login",
			Method: http.MethodPost,
			Path:   "api/auth/login",
			RequestBody: map[string]interface{}{
				"email":    "adhitya.ramadhanus@gmail.com",
				"password": "test123",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test forgotPassword",
			Method: http.MethodPost,
			Path:   "api/auth/password/forgot",
			RequestBody: map[string]interface{}{
				"email": "adhitya.ramadhanus@gmail.com",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test resetPassword",
			Method: http.MethodPost,
			Path:   "api/auth/password/reset",
			RequestBody: map[string]interface{}{
				"token":              "asdasdasdasdasdasdasd",
				"password":           "test123",
				"password_confirmed": "test123",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test verifyTFA",
			Method: http.MethodPost,
			Path:   "api/auth/tfa/verify",
			RequestBody: map[string]interface{}{
				"code": "123123",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test verifyTFABypass",
			Method: http.MethodPost,
			Path:   "api/auth/tfa/bypass",
			RequestBody: map[string]interface{}{
				"secret": "asdasdasdasdasdasdasd",
				"code":   "123123",
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", ts.URL, testCase.Path)
			req, _ := createHttpJSONRequest(testCase.Method, url, testCase.RequestBody)
			res, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			var decodedResponse map[string]interface{}
			json.NewDecoder(res.Body).Decode(&decodedResponse)
			fmt.Println(decodedResponse)
			assert.Equal(t, res.StatusCode, testCase.ExpectedStatus)
		})
	}
}
