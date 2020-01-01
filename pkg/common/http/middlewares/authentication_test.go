//+build unit

package middlewares_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/repository"
)

func TestTokenAuth(t *testing.T) {
	user := userland.User{
		Fullname: "Adhitya Ramadhanus",
		Email:    "adhitya.ramadhanus@gmail.com",
		ID:       1,
	}
	accessToken, err := security.CreateAccessToken(user, "jwtsecret_test", security.AccessTokenOptions{
		Expiration: security.UserAccessTokenExpiration,
		Scope:      security.UserTokenScope,
	})
	if err != nil {
		t.Fatalf("security.CreateAccessToken() err = %v; want nil", err)
	}

	keyValueService := repository.KeyValueService{}
	keyValueService.On("Get", keygenerator.TokenKey(accessToken.Key)).Return([]byte(accessToken.Value), nil)
	keyValueService.On("Get", keygenerator.TokenKey("test")).Return(nil, userland.ErrKeyNotFound)

	type args struct {
		authHeader string
	}
	testCases := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "invalid basic auth",
			args: args{
				authHeader: "Basic asdasdasd",
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "valid bearer auth",
			args: args{
				authHeader: fmt.Sprintf("Bearer %s", accessToken.Key),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "expired bearer auth",
			args: args{
				authHeader: "Bearer test",
			},
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authenticator := middlewares.TokenAuth(&keyValueService, "jwtsecret_test")
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}
			mw := authenticator(http.HandlerFunc(handler))

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatalf("http.NewRequest() err = %v; want nil", err)
			}
			req.Header.Set("Authorization", tc.args.authHeader)
			res := httptest.NewRecorder()

			mw.ServeHTTP(res, req)
			statusCode := res.Result().StatusCode
			if statusCode != tc.wantStatusCode {
				body, _ := ioutil.ReadAll(res.Result().Body)
				defer res.Result().Body.Close()
				t.Logf("response %s\n", string(body))
				t.Errorf("middlewares.TokenAuth() res.StatusCode = %d; want %d", statusCode, tc.wantStatusCode)
			}
		})
	}
}

func TestBasicAuth(t *testing.T) {
	username := "test"
	password := "coba"
	type args struct {
		authHeader string
	}
	testCases := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "invalid basic auth",
			args: args{
				authHeader: "Basic asdasdasd",
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "expired bearer auth",
			args: args{
				authHeader: "Bearer test",
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "valid basic auth",
			args: args{
				authHeader: fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))),
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authenticator := middlewares.BasicAuth(username, password)
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}
			mw := authenticator(http.HandlerFunc(handler))

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatalf("http.NewRequest() err = %v; want nil", err)
			}
			req.Header.Set("Authorization", tc.args.authHeader)
			res := httptest.NewRecorder()

			mw.ServeHTTP(res, req)
			statusCode := res.Result().StatusCode
			if statusCode != tc.wantStatusCode {
				body, _ := ioutil.ReadAll(res.Result().Body)
				defer res.Result().Body.Close()
				t.Logf("response %s\n", string(body))
				t.Errorf("middlewares.TokenAuth() res.StatusCode = %d; want %d", statusCode, tc.wantStatusCode)
			}
		})
	}
}
