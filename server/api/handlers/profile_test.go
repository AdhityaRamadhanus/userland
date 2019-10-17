//+build unit

package handlers_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_http "github.com/AdhityaRamadhanus/userland/common/http"
	"github.com/AdhityaRamadhanus/userland/mocks/middlewares"
	"github.com/AdhityaRamadhanus/userland/mocks/service/event"
	"github.com/AdhityaRamadhanus/userland/mocks/service/profile"
	"github.com/AdhityaRamadhanus/userland/server/api/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestProfileHandler(t *testing.T) {
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

	testCases := []struct {
		Name           string
		Method         string
		Path           string
		RequestBody    map[string]interface{}
		ExpectedStatus int
	}{
		{
			Name:           "Test getProfile",
			Method:         http.MethodGet,
			Path:           "api/me",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test setProfile",
			Method: http.MethodPost,
			Path:   "api/me",
			RequestBody: map[string]interface{}{
				"fullname": "Adhitya ramadhanus",
				"web":      "test",
				"location": "test",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test deletePicture",
			Method:         http.MethodDelete,
			Path:           "api/me/picture",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test getEmail",
			Method:         http.MethodGet,
			Path:           "api/me/email",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test requestChangeEmail",
			Method: http.MethodPost,
			Path:   "api/me/email",
			RequestBody: map[string]interface{}{
				"email": "adhitya.ramadhanus@gmail.com",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test changeEmail",
			Method: http.MethodPatch,
			Path:   "api/me/email",
			RequestBody: map[string]interface{}{
				"token": "asdasdasdasdas",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test changePassword",
			Method: http.MethodPost,
			Path:   "api/me/password",
			RequestBody: map[string]interface{}{
				"password_current":   "asdasdasdasdas",
				"password":           "asdasdasdasdas",
				"password_confirmed": "asdasdasdasdas",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test getTFAStatus",
			Method:         http.MethodGet,
			Path:           "api/me/tfa",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test enrollTFA",
			Method:         http.MethodGet,
			Path:           "api/me/tfa/enroll",
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test activateTFA",
			Method: http.MethodPost,
			Path:   "api/me/tfa/enroll",
			RequestBody: map[string]interface{}{
				"code":   "123123",
				"secret": "asdasdasdas",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test removeTFA",
			Method: http.MethodPost,
			Path:   "api/me/tfa/remove",
			RequestBody: map[string]interface{}{
				"password": "test123",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "Test deleteAccount",
			Method: http.MethodPost,
			Path:   "api/me/delete",
			RequestBody: map[string]interface{}{
				"password": "test123",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Test getEvents",
			Method:         http.MethodGet,
			Path:           "api/me/events",
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

func TestSetPicture(t *testing.T) {
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

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "profile.jpeg")

	file, err := os.Open("../../../test_files/1x1.jpg")
	defer file.Close()
	io.Copy(fileWriter, file)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	url := fmt.Sprintf("%s/%s", ts.URL, "api/me/picture")
	res, err := http.Post(url, contentType, bodyBuf)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)
}
