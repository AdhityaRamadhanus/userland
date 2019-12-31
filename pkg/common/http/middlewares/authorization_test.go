//+build unit

package middlewares_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	jwt "github.com/dgrijalva/jwt-go"
)

func TestAuthorization(t *testing.T) {
	user := userland.User{
		Fullname: "Adhitya Ramadhanus",
		Email:    "adhitya.ramadhanus@gmail.com",
		ID:       1,
	}
	jwtSecret := "jwtsecret_test"
	userAccessToken, _ := security.CreateAccessToken(user, jwtSecret, security.AccessTokenOptions{
		Expiration: security.UserAccessTokenExpiration,
		Scope:      security.UserTokenScope,
	})
	tfaAccessToken, _ := security.CreateAccessToken(user, jwtSecret, security.AccessTokenOptions{
		Expiration: security.TFATokenExpiration,
		Scope:      security.TFATokenScope,
	})

	type args struct {
		accessToken security.AccessToken
	}
	testCases := []struct {
		name           string
		args           args
		wantScope      string
		wantStatusCode int
	}{
		{
			name: "success",
			args: args{
				accessToken: userAccessToken,
			},
			wantScope:      security.UserTokenScope,
			wantStatusCode: http.StatusOK,
		},
		{
			name: "forbidden",
			args: args{
				accessToken: tfaAccessToken,
			},
			wantScope:      security.UserTokenScope,
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jwtToken, err := jwt.Parse(tc.args.accessToken.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil {
				t.Fatalf("jwt.Parse() err = %v; want nil", err)
			}

			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatalf("failed to assert jwtToken")
			}

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatalf("http.NewRequest() err = %v; want nil", err)
			}
			req = req.WithContext(context.WithValue(req.Context(), contextkey.AccessToken, map[string]interface{}(claims)))
			res := httptest.NewRecorder()

			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}
			authorize := middlewares.Authorize
			mw := authorize(http.HandlerFunc(handler), tc.wantScope)

			mw.ServeHTTP(res, req)
			defer res.Result().Body.Close()
			statusCode := res.Result().StatusCode
			if statusCode != tc.wantStatusCode {
				body, _ := ioutil.ReadAll(res.Result().Body)
				t.Logf("response %s\n", string(body))
				t.Errorf("middlewares.Authorize() res.StatusCode = %d; want %d", statusCode, tc.wantStatusCode)
			}
		})
	}
}
