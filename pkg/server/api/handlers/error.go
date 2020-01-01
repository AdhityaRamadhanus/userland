package handlers

import (
	"fmt"
	"net/http"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
	"github.com/AdhityaRamadhanus/userland/pkg/service/authentication"
	"github.com/AdhityaRamadhanus/userland/pkg/service/profile"
	"github.com/sirupsen/logrus"
)

var (
	ServiceErrorsHTTPMapping = map[error]struct {
		HTTPCode int
		ErrCode  string
	}{
		userland.ErrUserNotFound: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrUserNotFound",
		},
		authentication.ErrUserRegistered: {
			HTTPCode: http.StatusBadRequest,
			ErrCode:  "ErrUserRegistered",
		},
		authentication.ErrWrongOTP: {
			HTTPCode: http.StatusBadRequest,
			ErrCode:  "ErrWrongOTP",
		},
		authentication.ErrWrongPassword: {
			HTTPCode: http.StatusBadRequest,
			ErrCode:  "ErrWrongPassword",
		},
		authentication.ErrUserNotVerified: {
			HTTPCode: http.StatusBadRequest,
			ErrCode:  "ErrUserNotVerified",
		},
		authentication.ErrOTPInvalid: {
			HTTPCode: http.StatusBadRequest,
			ErrCode:  "ErrOTPInvalid",
		},
		profile.ErrWrongOTP: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrWrongOTP",
		},
		profile.ErrWrongPassword: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrWrongPassword",
		},
		profile.ErrEmailAlreadyUsed: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrEmailAlreadyUsed",
		},
	}
)

func handleServiceError(res http.ResponseWriter, req *http.Request, err error) {
	errorMapping, isErrorMapped := ServiceErrorsHTTPMapping[err]
	if isErrorMapped {
		render.JSON(res, errorMapping.HTTPCode, map[string]interface{}{
			"status": errorMapping.HTTPCode,
			"error": map[string]interface{}{
				"code":    errorMapping.ErrCode,
				"message": err.Error(),
			},
		})
		return
	}

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
			"message": "userland api unable to process the request",
		},
	})
	return
}
