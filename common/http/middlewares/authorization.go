package middlewares

import (
	"fmt"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/common/http/render"
)

func Authorize(next http.Handler, args ...interface{}) http.Handler {
	desiredTokenScope := args[0].(string)
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})

		if accessToken["scope"].(string) != desiredTokenScope { // invalid header
			render.JSON(res, http.StatusForbidden, map[string]interface{}{
				"status": http.StatusForbidden,
				"error": map[string]interface{}{
					"code":    "ErrForbiddenScope",
					"message": fmt.Sprintf("Use %s token", desiredTokenScope),
				},
			})
			return
		}

		next.ServeHTTP(res, req)
	})
}
