package middlewares

import "net/http"

type Middleware func(next http.Handler) http.Handler
type MiddlewareWithArgs func(next http.Handler, args ...interface{}) http.Handler
