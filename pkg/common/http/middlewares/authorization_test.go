//+build unit

package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestAuthorization(t *testing.T) {
	user := userland.User{
		Fullname: "Adhitya Ramadhanus",
		Email:    "adhitya.ramadhanus@gmail.com",
		ID:       1,
	}
	userAccessToken, _ := security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: security.UserAccessTokenExpiration,
		Scope:      security.UserTokenScope,
	})
	tfaAccessToken, _ := security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: security.TFATokenExpiration,
		Scope:      security.TFATokenScope,
	})

	testCases := []struct {
		AccessToken        security.AccessToken
		DesiredScope       string
		ExpectedStatusCode int
	}{
		{
			AccessToken:        userAccessToken,
			DesiredScope:       security.UserTokenScope,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			AccessToken:        tfaAccessToken,
			DesiredScope:       security.UserTokenScope,
			ExpectedStatusCode: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		jwtToken, _ := jwt.Parse(testCase.AccessToken.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		contextSetter := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, _ := jwtToken.Claims.(jwt.MapClaims)
				r = r.WithContext(context.WithValue(r.Context(), contextkey.AccessToken, map[string]interface{}(claims)))
				next.ServeHTTP(w, r)
			})
		}
		authorize := middlewares.Authorize
		defaultHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		ts := httptest.NewServer(contextSetter(authorize(defaultHandler, testCase.DesiredScope)))
		defer ts.Close()

		res, err := http.Get(ts.URL)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, testCase.ExpectedStatusCode)
	}
}
