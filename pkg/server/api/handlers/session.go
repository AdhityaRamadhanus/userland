package handlers

import (
	"net/http"

	"github.com/AdhityaRamadhanus/userland/pkg/server/api/serializers"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/service/session"
	"github.com/gorilla/mux"
)

type SessionHandler struct {
	Authorization  middlewares.MiddlewareWithArgs
	Authenticator  middlewares.Middleware
	SessionService session.Service
	ProfileService profile.Service
}

func (h SessionHandler) RegisterRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/api").Subrouter()

	authenticate := h.Authenticator
	authorize := h.Authorization

	listSession := authenticate(authorize(http.HandlerFunc(h.listSession), security.UserTokenScope))
	endCurrentSession := authenticate(authorize(http.HandlerFunc(h.endCurrentSession), security.UserTokenScope))
	endOtherSession := authenticate(authorize(http.HandlerFunc(h.endOtherSession), security.UserTokenScope))
	createRefreshToken := authenticate(authorize(http.HandlerFunc(h.createRefreshToken), security.UserTokenScope))
	createNewAccessToken := authenticate(authorize(http.HandlerFunc(h.createNewAccessToken), security.RefreshTokenScope))

	subRouter.Handle("/me/session", listSession).Methods("GET")
	subRouter.Handle("/me/session", endCurrentSession).Methods("DELETE")
	subRouter.Handle("/me/session/other", endOtherSession).Methods("DELETE")
	subRouter.Handle("/me/session/refresh_token", createRefreshToken).Methods("GET")
	subRouter.Handle("/me/session/access_token", createNewAccessToken).Methods("GET")
}

func (h SessionHandler) listSession(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	accessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(accessToken["userid"].(float64))

	sessions, err := h.SessionService.ListSession(userID)
	if err != nil {
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
		return
	}

	refreshToken, err := h.SessionService.CreateRefreshToken(user, accessTokenKey)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"access_token": serializers.SerializeAccessTokenToJSON(refreshToken),
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
		handleServiceError(res, req, err)
		return
	}

	// create access token
	accessToken, err := h.SessionService.CreateNewAccessToken(user, refreshTokenKey)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}
	// create session
	h.SessionService.CreateSession(user.ID, userland.Session{
		ID:         accessToken.Key,
		Token:      accessToken.Value,
		IP:         clientInfo["ip"].(string),
		ClientID:   clientInfo["client_id"].(int),
		ClientName: clientInfo["client_name"].(string),
		Expiration: security.UserAccessTokenExpiration,
	})
	// delete prev session
	h.SessionService.EndSession(user.ID, prevSessionID)
	render.JSON(res, http.StatusOK, map[string]interface{}{
		"access_token": serializers.SerializeAccessTokenToJSON(accessToken),
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
