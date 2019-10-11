package handlers

import (
	"net/http"

	"github.com/AdhityaRamadhanus/userland/server/api/serializers"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/common/http/render"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/service/profile"
	"github.com/AdhityaRamadhanus/userland/service/session"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SessionHandler struct {
	Authenticator  *middlewares.Authenticator
	SessionService session.Service
	ProfileService profile.Service
}

func (h SessionHandler) RegisterRoutes(router *mux.Router) {
	authenticate := h.Authenticator.Authenticate
	userAuthorize := middlewares.Authorize(security.UserTokenScope)
	refreshAuthorize := middlewares.Authorize(security.RefreshTokenScope)

	router.HandleFunc("/me/session", authenticate(userAuthorize(h.listSession))).Methods("GET")
	router.HandleFunc("/me/session", authenticate(userAuthorize(h.endCurrentSession))).Methods("DELETE")
	router.HandleFunc("/me/session/other", authenticate(userAuthorize(h.endOtherSession))).Methods("DELETE")
	router.HandleFunc("/me/session/refresh_token", authenticate(userAuthorize(h.createRefreshToken))).Methods("GET")
	router.HandleFunc("/me/session/access_token", authenticate(refreshAuthorize(h.createNewAccessToken))).Methods("GET")
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

func (h SessionHandler) createRefreshToken(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	accessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	refreshToken, err := h.SessionService.CreateRefreshToken(user, accessTokenKey)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"access_token": map[string]interface{}{
			"value":      refreshToken.Key,
			"type":       refreshToken.Type,
			"expired_at": refreshToken.ExpiredAt,
		},
	})
}

func (h SessionHandler) createNewAccessToken(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	refreshToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	refreshTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(refreshToken["userid"].(float64))
	prevSessionID := refreshToken["previous_session_id"].(string)

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// create access token
	accessToken, err := h.SessionService.CreateNewAccessToken(user, refreshTokenKey)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}
	// create session
	h.SessionService.CreateSession(user.ID, userland.Session{
		ID:         accessToken.Key,
		Token:      accessToken.Value,
		IP:         clientInfo["ip"].(string),
		ClientID:   clientInfo["client_id"].(int),
		ClientName: clientInfo["client_name"].(string),
	})
	// delete prev session
	h.SessionService.EndSession(user.ID, prevSessionID)
	render.JSON(res, http.StatusOK, map[string]interface{}{
		"access_token": map[string]interface{}{
			"value":      accessToken.Key,
			"type":       accessToken.Type,
			"expired_at": accessToken.ExpiredAt,
		},
	})
}

func serializeSessions(sessions userland.Sessions, currentSessionID string) []map[string]interface{} {
	serializedSessions := []map[string]interface{}{}
	for _, _session := range sessions {
		serializedSession := serializers.SerializeSessionToJSON(_session)
		if _session.ID == currentSessionID {
			serializedSession["is_current"] = true
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