package middlewares

import (
	"context"
	b64 "encoding/base64"

	"github.com/go-errors/errors"

	"net/http"
	"strings"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	jwt "github.com/dgrijalva/jwt-go"
)

func parseAuthorizationHeader(authHeader, scheme string) (cred string, err error) {
	splittedHeader := strings.Split(authHeader, " ")
	if len(splittedHeader) != 2 {
		return "", errors.New("Cannot parse authorization header")
	}
	parsedScheme := splittedHeader[0]
	if scheme != parsedScheme {
		return "", errors.New("Unexpected Scheme, expected " + scheme)
	}
	return splittedHeader[1], nil
}

//Authenticate request
func TokenAuth(keyValueService userland.KeyValueService, jwtSecret string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			authHeader, ok := req.Header["Authorization"]
			if !ok || len(authHeader) == 0 { // invalid header
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidAuthorizationHeader",
						"message": "Authorization Header is not present",
					},
				})
				return
			}

			cred, err := parseAuthorizationHeader(authHeader[0], "Bearer")
			if err != nil {
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidAuthorizationHeader",
						"message": err.Error(),
					},
				})
				return
			}

			token, err := keyValueService.Get(keygenerator.TokenKey(cred))
			if err != nil {
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidAccessToken",
						"message": "Token is expired/not found",
					},
				})
				return
			}

			jwtToken, err := jwt.Parse(string(token), func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("Unexpected signing method")
				}
				// TODO move to config
				return []byte(jwtSecret), nil
			})
			if err != nil {
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidAccessToken",
						"message": err.Error(),
					},
				})
				return
			}

			claims, _ := jwtToken.Claims.(jwt.MapClaims)
			req = req.WithContext(context.WithValue(req.Context(), contextkey.AccessToken, map[string]interface{}(claims)))
			req = req.WithContext(context.WithValue(req.Context(), contextkey.AccessTokenKey, cred))
			next.ServeHTTP(res, req)
		})
	}
}

func BasicAuth(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			authHeader, ok := req.Header["Authorization"]
			if !ok || len(authHeader) == 0 { // invalid header
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidAuthorizationHeader",
						"message": "Authorization Header is not present",
					},
				})
				return
			}

			cred, err := parseAuthorizationHeader(authHeader[0], "Basic")
			if err != nil {
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidAuthorizationHeader",
						"message": err.Error(),
					},
				})
				return
			}

			sDec, _ := b64.StdEncoding.DecodeString(cred)
			splitDec := strings.Split(string(sDec), ":")

			if len(splitDec) < 2 {
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidBasicAuth",
						"message": "username/password is wrong",
					},
				})
				return
			}
			if username != splitDec[0] || password != splitDec[1] {
				render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
					"status": http.StatusUnauthorized,
					"error": map[string]interface{}{
						"code":    "ErrInvalidBasicAuth",
						"message": "username/password is wrong",
					},
				})
				return
			}

			next.ServeHTTP(res, req)
		})
	}
}
