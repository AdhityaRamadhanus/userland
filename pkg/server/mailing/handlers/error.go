package handlers

import (
	"fmt"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
	"github.com/sirupsen/logrus"
)

func handleServiceError(res http.ResponseWriter, req *http.Request, err error) {
	logrus.WithFields(logrus.Fields{
		"method":       req.Method,
		"route":        req.URL.Path,
		"status_code":  http.StatusInternalServerError,
		"stack_trace":  fmt.Sprintf("%+v", err),
		"client":       req.Header.Get("X-API-ClientID"),
		"x-request-id": req.Header.Get("X-Request-ID"),
	}).WithError(err).Error("Error Handler")

	render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrInternalServer",
			"message": "userland mail api unable to process the request",
		},
	})
	return
}
