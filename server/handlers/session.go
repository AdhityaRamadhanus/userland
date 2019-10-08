package handlers

import (
	"net/http"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/userland/server/middlewares"
	"github.com/AdhityaRamadhanus/userland/server/render"
	"github.com/AdhityaRamadhanus/userland/session"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SessionHandler struct {
	Authenticator  *middlewares.Authenticator
	SessionService session.Service
}

func (h SessionHandler) RegisterRoutes(router *mux.Router) {
	authenticate := h.Authenticator.Authenticate
	authorize := middlewares.Authorize

	router.HandleFunc("/me/session", authenticate(authorize(security.UserTokenScope, h.listSession))).Methods("GET")
	router.HandleFunc("/me/session", authenticate(authorize(security.UserTokenScope, h.endCurrentSession))).Methods("DELETE")
	router.HandleFunc("/me/session/other", authenticate(authorize(security.UserTokenScope, h.endOtherSession))).Methods("DELETE")
}

func (h SessionHandler) listSession(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	accessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(accessToken["userid"].(float64))

	sessions, err := h.SessionService.ListSession(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"data": serializeSessions(sessions, accessTokenKey)})
}

func (h SessionHandler) endCurrentSession(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	accessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(accessToken["userid"].(float64))

	err := h.SessionService.EndSession(userID, accessTokenKey)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h SessionHandler) endOtherSession(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	accessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(accessToken["userid"].(float64))

	err := h.SessionService.EndOtherSessions(userID, accessTokenKey)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func serializeSessions(sessions userland.Sessions, currentSessionID string) []map[string]interface{} {
	serializedSessions := []map[string]interface{}{}
	for _, _session := range sessions {
		serializedSession := map[string]interface{}{
			"session_id": _session.ID,
			"is_current": _session.ID == currentSessionID,
			"ip":         _session.IP,
			"client": map[string]interface{}{
				"id":   _session.ClientID,
				"name": _session.ClientName,
			},
			"created_at": _session.CreatedAt,
			"updated_at": _session.UpdatedAt,
		}

		serializedSessions = append(serializedSessions, serializedSession)
	}

	return serializedSessions
}

func (h SessionHandler) handleServiceError(res http.ResponseWriter, req *http.Request, err error) {
	ServiceErrorsHTTPMapping := map[error]struct {
		HTTPCode int
		ErrCode  string
	}{
		userland.ErrUserNotFound: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrUserNotFound",
		},
	}

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

	log.WithFields(log.Fields{
		"endpoint":     req.URL.Path,
		"client":       req.Header.Get("X-API-ClientID"),
		"x-request-id": req.Header.Get("X-Request-ID"),
	}).WithError(err).Error("Error Session Handler")

	render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrInternalServer",
			"message": "userland api unable to process the request",
		},
	})
	return
}
