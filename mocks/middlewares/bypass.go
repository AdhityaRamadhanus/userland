package middlewares

import (
	"net/http"
)

var (
	Bypass = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(res, req)
		})
	}

	BypassWithArgs = func(next http.Handler, args ...interface{}) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(res, req)
		})
	}
)
