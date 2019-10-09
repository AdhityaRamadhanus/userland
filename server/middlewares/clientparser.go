package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/AdhityaRamadhanus/userland/server/internal/contextkey"
)

func getClientIP(req *http.Request) string {
	ipAddress, xRealIPPresent := req.Header["X-Real-Ip"]
	if xRealIPPresent {
		return ipAddress[0]
	}

	ipAddress, xForwardedForPresent := req.Header["X-Forwarded-For"]
	if xForwardedForPresent {
		return ipAddress[0]
	}

	return req.RemoteAddr
}

//LogRequest with info level every http request, unless production
func ParseClientInfo(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		clientInfoMap := map[string]interface{}{
			"client_id":   -1,
			"client_name": "unknown",
			"ip":          getClientIP(req),
		}
		clientInfo, clientInfoPresent := req.Header["X-API-ClientID"]
		if clientInfoPresent {
			// assume clientInfo is in "<name>:<id>" format
			clientInfoSplitted := strings.Split(clientInfo[0], ":")
			clientInfoMap["client_id"], _ = strconv.Atoi(clientInfoSplitted[1])
			clientInfoMap["client_name"] = clientInfoSplitted[0]
		}

		// get IP address
		req = req.WithContext(context.WithValue(req.Context(), contextkey.ClientInfo, map[string]interface{}(clientInfoMap)))
		nextHandler.ServeHTTP(res, req)
	})
}
