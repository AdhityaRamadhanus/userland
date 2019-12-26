//+build unit

package middlewares_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/mocks/repository"
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

	keyValueService := repository.KeyValueService{}
	keyValueService.On("Get", keygenerator.TokenKey(accessToken.Key)).Return([]byte(accessToken.Value), nil)
	keyValueService.On("Get", keygenerator.TokenKey("test")).Return(nil, userland.ErrKeyNotFound)

	for _, testCase := range testCases {

		authenticator := middlewares.TokenAuth(&keyValueService)
		ts := httptest.NewServer(authenticator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
