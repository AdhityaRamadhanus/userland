package middlewares

import (
	"context"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
)

var (
	ClientParser = func(next http.Handler, args ...interface{}) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			clientInfoMap := map[string]interface{}{
				"client_id":   -1,
				"client_name": "unknown",
				"ip":          "test",
				"user_agent":  "test",
			}

			req = req.WithContext(context.WithValue(req.Context(), contextkey.ClientInfo, map[string]interface{}(clientInfoMap)))
			next.ServeHTTP(res, req)
		})
	}
)
