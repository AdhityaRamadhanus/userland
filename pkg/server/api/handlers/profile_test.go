//+build unit

package handlers_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_http "github.com/AdhityaRamadhanus/userland/pkg/common/http"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/event"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/server/api/handlers"
	"github.com/gorilla/mux"
)

func TestProfileHandler_inputValidation(t *testing.T) {
	profileService := profile.SimpleProfileService{CalledMethods: map[string]bool{}}
	eventService := event.SimpleEventService{CalledMethods: map[string]bool{}}

	profileHandler := handlers.ProfileHandler{
		RateLimiter:    middlewares.BypassWithArgs,
		Authorization:  middlewares.BypassWithArgs,
		Authenticator:  middlewares.Authentication,
		ProfileService: profileService,
		EventService:   eventService,
	}
	router := mux.NewRouter().StrictSlash(true)
	profileHandler.RegisterRoutes(router)

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
			name: "POST api/me",
			args: args{
				method: http.MethodPost,
				path:   "api/me",
				requestBody: map[string]interface{}{
					"fullname": "adhitya ramadhanus",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "DELETE api/me/picture",
			args: args{
				method:      http.MethodDelete,
				path:        "api/me/picture",
				requestBody: map[string]interface{}{},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/me/email",
			args: args{
				method: http.MethodPost,
				path:   "api/me/email",
				requestBody: map[string]interface{}{
					"email": "adhitya.ramadhanus@gmail.com",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "PATCH api/me/email",
			args: args{
				method: http.MethodPatch,
				path:   "api/me/email",
				requestBody: map[string]interface{}{
					"token": "asdasdasdasdas",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/me/password",
			args: args{
				method: http.MethodPost,
				path:   "api/me/password",
				requestBody: map[string]interface{}{
					"password_current":   "asdasdasdasdas",
					"password":           "asdasdasdasdas",
					"password_confirmed": "asdasdasdasdas",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/me/tfa/enroll",
			args: args{
				method: http.MethodPost,
				path:   "api/me/tfa/enroll",
				requestBody: map[string]interface{}{
					"code":   "123123",
					"secret": "asdasdasdas",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "POST api/me/tfa/remove",
			args: args{
				method: http.MethodPost,
				path:   "api/me/tfa/remove",
				requestBody: map[string]interface{}{
					"password": "123123",
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "DELETE api/me/delete",
			args: args{
				method: http.MethodDelete,
				path:   "api/me/delete",
				requestBody: map[string]interface{}{
					"password": "123123",
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

func TestProfileHandler_setPictureValidation(t *testing.T) {
	profileService := profile.SimpleProfileService{CalledMethods: map[string]bool{}}
	eventService := event.SimpleEventService{CalledMethods: map[string]bool{}}

	profileHandler := handlers.ProfileHandler{
		RateLimiter:    middlewares.BypassWithArgs,
		Authorization:  middlewares.BypassWithArgs,
		Authenticator:  middlewares.Authentication,
		ProfileService: profileService,
		EventService:   eventService,
	}
	router := mux.NewRouter().StrictSlash(true)
	profileHandler.RegisterRoutes(router)

	ts := httptest.NewServer(middlewares.ClientParser(router))
	defer ts.Close()

	type args struct {
		filePath string
	}
	testCases := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "1x1.jpg",
			args: args{
				filePath: "./testdata/1x1.jpg",
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)
			fileWriter, err := bodyWriter.CreateFormFile("file", "profile.jpeg")
			if err != nil {
				t.Fatalf("bodyWriter.CreateFormFile(file, profile.jpeg) err = %v; want nil", err)
			}

			file, err := os.Open(tc.args.filePath)
			if err != nil {
				t.Fatalf("os.Open(path) err = %v; want nil", err)
			}
			defer file.Close()
			io.Copy(fileWriter, file)

			contentType := bodyWriter.FormDataContentType()
			bodyWriter.Close()

			url := fmt.Sprintf("%s/%s", ts.URL, "api/me/picture")
			res, err := http.Post(url, contentType, bodyBuf)
			if err != nil {
				t.Fatalf("http.Post() err = %v; want nil", err)
			}

			statusCode := res.StatusCode
			if statusCode != tc.wantStatusCode {
				body, _ := ioutil.ReadAll(res.Body)
				defer res.Body.Close()
				t.Logf("response %s\n", string(body))
				t.Errorf("res.StatusCode = %d; want %d", statusCode, tc.wantStatusCode)
			}
		})
	}
}
