package middlewares

import (
	b64 "encoding/base64"
	"net/http"
	"strings"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
)

//Authenticate request
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
