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
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/session"
	"github.com/AdhityaRamadhanus/userland/pkg/server/api/handlers"
	"github.com/gorilla/mux"
)

func TestSessionHandler_inputValidation(t *testing.T) {
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
			name: "DELETE api/me/session",
			args: args{
				method:      http.MethodDelete,
				path:        "api/me/session",
				requestBody: map[string]interface{}{},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "DELETE api/me/session/other",
			args: args{
				method:      http.MethodDelete,
				path:        "api/me/session/other",
				requestBody: map[string]interface{}{},
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
