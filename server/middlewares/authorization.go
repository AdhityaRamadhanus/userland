package middlewares

import (
	"fmt"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/userland/server/render"
)

func Authorize(desiredTokenScope string, nextHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})

		if accessToken["scope"].(string) != desiredTokenScope { // invalid header
			render.JSON(res, http.StatusUnauthorized, map[string]interface{}{
				"status": http.StatusForbidden,
				"error": map[string]interface{}{
					"code":    "ErrForbiddenScope",
					"message": fmt.Sprintf("Use %s token", desiredTokenScope),
				},
			})
			return
		}

		nextHandler(res, req)
	})
}
