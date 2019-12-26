package middlewares

import (
	"context"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
)

var (
	Authentication = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			claims := map[string]interface{}{
				"userid":   float64(1),
				"fullname": "unknown",
				"email":    "test",
				"scope":    "test",
			}

			req = req.WithContext(context.WithValue(req.Context(), contextkey.AccessToken, claims))
			req = req.WithContext(context.WithValue(req.Context(), contextkey.AccessTokenKey, "test"))
			next.ServeHTTP(res, req)
		})
	}

	AuthenticationWithCustomClaims = func(customClaims map[string]interface{}) func(next http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				claims := map[string]interface{}{
					"userid":   float64(1),
					"fullname": "unknown",
					"email":    "test",
				}

				for key, val := range customClaims {
					claims[key] = val
				}

				req = req.WithContext(context.WithValue(req.Context(), contextkey.AccessToken, claims))
				req = req.WithContext(context.WithValue(req.Context(), contextkey.AccessTokenKey, "test"))
				next.ServeHTTP(res, req)
			})
		}
	}
)
