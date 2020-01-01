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
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/server/mailing/handlers"
	"github.com/gorilla/mux"
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
			name: "POST api/mail/otp",
			args: args{
				method: http.MethodPost,
				path:   "api/mail/otp",
				requestBody: map[string]interface{}{
					"recipient_name": "Adhitya Ramadhanus",
					"recipient":      "adhitya.ramadhanus@gmail.com",
					"type":           "TFA Verification",
					"otp":            "123123",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/mail/verification",
			args: args{
				method: http.MethodPost,
				path:   "api/mail/verification",
				requestBody: map[string]interface{}{
					"recipient_name":    "Adhitya Ramadhanus",
					"recipient":         "adhitya.ramadhanus@gmail.com",
					"verification_link": "testaja",
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
