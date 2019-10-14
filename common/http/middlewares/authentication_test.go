//+build unit

package middlewares_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuthentication(t *testing.T) {
	user := userland.User{
		Fullname: "Adhitya Ramadhanus",
		Email:    "adhitya.ramadhanus@gmail.com",
		ID:       1,
	}
	accessToken, _ := security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: security.UserAccessTokenExpiration,
		Scope:      security.UserTokenScope,
	})

	testCases := []struct {
		AuthHeader         string
		ExpectedStatusCode int
	}{
		{
			AuthHeader:         "Basic asdasd",
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			AuthHeader:         fmt.Sprintf("Bearer %s", accessToken.Key),
			ExpectedStatusCode: http.StatusOK,
		},
		{
			AuthHeader:         fmt.Sprintf("Bearer test"),
			ExpectedStatusCode: http.StatusUnauthorized,
		},
	}

	keyValueService := mocks.KeyValueService{}
	keyValueService.On("Get", keygenerator.TokenKey(accessToken.Key)).Return([]byte(accessToken.Value), nil)
	keyValueService.On("Get", keygenerator.TokenKey("test")).Return(nil, userland.ErrKeyNotFound)

	for _, testCase := range testCases {

		authenticator := middlewares.NewAuthenticator(&keyValueService)
		ts := httptest.NewServer(authenticator.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})))
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
		req.Header.Set("Authorization", testCase.AuthHeader)
		res, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, testCase.ExpectedStatusCode)
	}
}
