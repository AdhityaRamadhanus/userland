//+build unit

package handlers_test

import (
	"fmt"
	"io/ioutil"
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
)

func TestAuthenticationHandler_inputValidation(t *testing.T) {
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

	type args struct {
		path        string
		method      string
		requestBody map[string]interface{}
	}
	testCases := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "POST api/auth/register",
			args: args{
				method: http.MethodPost,
				path:   "api/auth/register",
				requestBody: map[string]interface{}{
					"email":              "adhitya.ramadhanus@gmail.com",
					"password":           "test123",
					"password_confirmed": "test123",
					"fullname":           "adhitya ramadhanus",
				},
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "POST api/auth/verification",
			args: args{
				method: http.MethodPost,
				path:   "api/auth/verification",
				requestBody: map[string]interface{}{
					"type":      "email.verify",
					"recipient": "adhitya.ramadhanus@gmail.com",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "PATCH api/auth/verification",
			args: args{
				method: http.MethodPatch,
				path:   "api/auth/verification",
				requestBody: map[string]interface{}{
					"verification_id": "asdasdasdasdasdasd",
					"code":            "test123",
					"email":           "adhitya.ramadhanus@gmail.com",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/auth/login",
			args: args{
				method: http.MethodPost,
				path:   "api/auth/login",
				requestBody: map[string]interface{}{
					"email":    "adhitya.ramadhanus@gmail.com",
					"password": "test123",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/auth/password/forgot",
			args: args{
				method: http.MethodPost,
				path:   "api/auth/password/forgot",
				requestBody: map[string]interface{}{
					"email": "adhitya.ramadhanus@gmail.com",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/auth/password/reset",
			args: args{
				method: http.MethodPost,
				path:   "api/auth/password/reset",
				requestBody: map[string]interface{}{
					"token":              "asdasdasdasdasdasdasd",
					"password":           "test123",
					"password_confirmed": "test123",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/auth/tfa/verify",
			args: args{
				method: http.MethodPost,
				path:   "api/auth/tfa/verify",
				requestBody: map[string]interface{}{
					"code": "123123",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/auth/tfa/bypass",
			args: args{
				method: http.MethodPost,
				path:   "api/auth/tfa/bypass",
				requestBody: map[string]interface{}{
					"secret": "asdasdasdasdasdasdasd",
					"code":   "123123",
				},
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", ts.URL, tc.args.path)
			req, err := _http.CreateJSONRequest(tc.args.method, url, tc.args.requestBody)
			if err != nil {
				t.Fatalf("_http.CreateJSONRequest() err = %v; want nil", err)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("http.DefaultClient.Do() err = %v; want nil", err)
			}
			statusCode := res.StatusCode
			if statusCode != tc.wantStatusCode {
				body, _ := ioutil.ReadAll(res.Body)
				defer res.Body.Close()
				t.Logf("response %s\n", string(body))
				t.Errorf("%s res.StatusCode = %d; want %d", tc.args.path, statusCode, tc.wantStatusCode)
			}
		})
	}
}
