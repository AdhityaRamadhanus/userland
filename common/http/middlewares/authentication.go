package middlewares

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/common/http/render"
	jwt "github.com/dgrijalva/jwt-go"
)

type Authenticator struct {
	keyValueService userland.KeyValueService
}

func NewAuthenticator(keyValueService userland.KeyValueService) *Authenticator {
	return &Authenticator{
		keyValueService: keyValueService,
	}
}

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
func (a *Authenticator) Authenticate(nextHandler http.HandlerFunc) http.HandlerFunc {
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

		token, err := a.keyValueService.Get(keygenerator.TokenKey(cred))
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
			return []byte(os.Getenv("JWT_SECRET")), nil
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
		nextHandler(res, req)
	})
}
