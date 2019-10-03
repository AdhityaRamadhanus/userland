package middlewares

import (
	"net/http"

	"github.com/AdhityaRamadhanus/userland/common/security"
)

//TraceRequest add X-Request-ID header to request
func TraceRequest(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		req.Header.Set("X-Request-ID", security.GenerateUUID())
		res.Header().Set("X-Request-ID", security.GenerateUUID())

		nextHandler.ServeHTTP(res, req)
	})
}
