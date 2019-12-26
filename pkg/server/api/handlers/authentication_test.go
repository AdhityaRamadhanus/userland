//+build unit

package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_http "github.com/AdhityaRamadhanus/userland/pkg/common/http"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/authentication"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/event"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/session"
	"github.com/AdhityaRamadhanus/userland/pkg/server/api/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationHandler(t *testing.T) {
	profileService := profile.SimpleProfileService{CalledMethods: map[string]bool{}}
	sessionService := session.SimpleSessionService{CalledMethods: map[string]bool{}}
	authenticationService := authentication.SimpleAuthenticationService{CalledMethods: map[string]bool{}}
	eventService := event.SimpleEventService{CalledMethods: map[string]bool{}}

	authenticationHandler := handlers.AuthenticationHandler{
		RateLimiter:           middlewares.BypassWithArgs,
		Authorization:         middlewares.BypassWithArgs,
		Authenticator:         middlewares.Authentication,
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
			req, _ := _http.CreateJSONRequest(testCase.Method, url, testCase.RequestBody)
			res, err := http.DefaultClient.Do(req)
			assert.Nil(t, err)
			assert.Equal(t, res.StatusCode, testCase.ExpectedStatus)
		})
	}
}
