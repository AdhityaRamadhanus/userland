//+build unit

package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_http "github.com/AdhityaRamadhanus/userland/common/http"
	"github.com/AdhityaRamadhanus/userland/mocks/middlewares"
	"github.com/AdhityaRamadhanus/userland/mocks/service/profile"
	"github.com/AdhityaRamadhanus/userland/mocks/service/session"
	"github.com/AdhityaRamadhanus/userland/server/api/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestSessionHandler(t *testing.T) {
	profileService := profile.SimpleProfileService{CalledMethods: map[string]bool{}}
	sessionService := session.SimpleSessionService{CalledMethods: map[string]bool{}}

	tokenClaims := map[string]interface{}{
		"previous_session_id": "test",
	}
	sessionHandler := handlers.SessionHandler{
		Authorization:  middlewares.BypassWithArgs,
		Authenticator:  middlewares.AuthenticationWithCustomClaims(tokenClaims),
		ProfileService: profileService,
		SessionService: sessionService,
	}
	router := mux.NewRouter().StrictSlash(true)
	sessionHandler.RegisterRoutes(router)

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
			Name:           "Test listSession",
			Method:         http.MethodGet,
			Path:           "api/me/session",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test endCurrentSession",
			Method:         http.MethodDelete,
			Path:           "api/me/session",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test endOtherSession",
			Method:         http.MethodDelete,
			Path:           "api/me/session/other",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test createRefreshToken",
			Method:         http.MethodGet,
			Path:           "api/me/session/refresh_token",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test createNewAccessToken",
			Method:         http.MethodGet,
			Path:           "api/me/session/access_token",
			RequestBody:    map[string]interface{}{},
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
