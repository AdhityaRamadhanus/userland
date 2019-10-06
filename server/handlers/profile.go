package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/asaskevich/govalidator"

	"github.com/AdhityaRamadhanus/userland/profile"
	"github.com/AdhityaRamadhanus/userland/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/userland/server/middlewares"
	"github.com/AdhityaRamadhanus/userland/server/render"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type ProfileHandler struct {
	Authenticator  *middlewares.Authenticator
	ProfileService profile.Service
}

func (h ProfileHandler) RegisterRoutes(router *mux.Router) {
	authenticate := h.Authenticator.Authenticate
	authorize := middlewares.Authorize
	router.HandleFunc("/me", authenticate(authorize(security.UserTokenScope, h.getProfile))).Methods("GET")
	router.HandleFunc("/me", authenticate(authorize(security.UserTokenScope, h.updateProfile))).Methods("POST")
	router.HandleFunc("/me", authenticate(authorize(security.UserTokenScope, h.updateProfile))).Methods("POST")
	router.HandleFunc("/me/email", authenticate(authorize(security.UserTokenScope, h.getEmail))).Methods("GET")
	router.HandleFunc("/me/email", authenticate(authorize(security.UserTokenScope, h.requestChangeEmail))).Methods("POST")
	router.HandleFunc("/me/email", authenticate(authorize(security.UserTokenScope, h.changeEmail))).Methods("PATCH")
	router.HandleFunc("/me/password", authenticate(authorize(security.UserTokenScope, h.changePassword))).Methods("POST")
	router.HandleFunc("/me/tfa", authenticate(authorize(security.UserTokenScope, h.getTFAStatus))).Methods("GET")
	router.HandleFunc("/me/tfa/enroll", authenticate(authorize(security.UserTokenScope, h.enrollTFA))).Methods("GET")
	router.HandleFunc("/me/tfa/enroll", authenticate(authorize(security.UserTokenScope, h.activateTFA))).Methods("POST")
	router.HandleFunc("/me/tfa/remove", authenticate(authorize(security.UserTokenScope, h.removeTFA))).Methods("POST")
	router.HandleFunc("/me/delete", authenticate(authorize(security.UserTokenScope, h.deleteAccount))).Methods("POST")
}

func (h *ProfileHandler) getProfile(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Profile")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"fullname":   user.Fullname,
		"location":   user.Location,
		"bio":        user.Bio,
		"web":        user.WebURL,
		"picture":    user.PictureURL,
		"created_at": user.CreatedAt,
	})
}

func (h *ProfileHandler) updateProfile(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Profile")

		RenderError(res, ErrSomethingWrong)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	updateProfileRequest := struct {
		Fullname string `json:"fullname" valid:"required"`
		Bio      string `json:"bio" valid:"-"`
		Location string `json:"location" valid:"-"`
		Web      string `json:"web" valid:"-"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &updateProfileRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(updateProfileRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	user.Fullname = updateProfileRequest.Fullname
	if len(updateProfileRequest.Bio) > 0 {
		user.Bio = updateProfileRequest.Bio
	}
	if len(updateProfileRequest.Location) > 0 {
		user.Location = updateProfileRequest.Location
	}
	if len(updateProfileRequest.Web) > 0 {
		user.WebURL = updateProfileRequest.Web
	}

	if err = h.ProfileService.SetProfile(user); err != nil {
		log.WithFields(log.Fields{
			"request":      updateProfileRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Register User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *ProfileHandler) getEmail(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"email": user.Email,
	})
}

func (h *ProfileHandler) requestChangeEmail(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	changeEmailRequest := struct {
		Email string `json:"email" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changeEmailRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(changeEmailRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if _, err = h.ProfileService.RequestChangeEmail(user, changeEmailRequest.Email); err != nil {
		log.WithFields(log.Fields{
			"request":      changeEmailRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Register User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *ProfileHandler) changeEmail(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	changeEmailRequest := struct {
		Token string `json:"token" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changeEmailRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(changeEmailRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.ChangeEmail(user, changeEmailRequest.Token); err != nil {
		log.WithFields(log.Fields{
			"request":      changeEmailRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Change Email User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *ProfileHandler) changePassword(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	changePasswordRequest := struct {
		CurrentPassword      string `json:"password_current" valid:"required"`
		NewPassword          string `json:"password" valid:"required"`
		NewConfirmedPassword string `json:"password_confirmed" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changePasswordRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(changePasswordRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.ChangePassword(user, changePasswordRequest.CurrentPassword, changePasswordRequest.NewPassword); err != nil {
		log.WithFields(log.Fields{
			"request":      changePasswordRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Change Email User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *ProfileHandler) getTFAStatus(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	response := map[string]interface{}{
		"enabled_at": "",
	}
	if user.TFAEnabled {
		response["enabled"] = true
		response["enabled_at"] = user.TFAEnabledAt
	}
	render.JSON(res, http.StatusOK, response)
}

func (h *ProfileHandler) enrollTFA(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	secret, qrCodeImageBase64, err := h.ProfileService.EnrollTFA(user)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"secret": secret,
		"qr":     fmt.Sprintf(`data:image/png;base64,%s`, qrCodeImageBase64),
	})
}

func (h *ProfileHandler) activateTFA(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	activateTFARequest := struct {
		Secret string `json:"secret" valid:"required"`
		Code   string `json:"code" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &activateTFARequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(activateTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	backupCodes, err := h.ProfileService.ActivateTFA(user, activateTFARequest.Secret, activateTFARequest.Code)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"backup_codes": backupCodes,
	})
}

func (h *ProfileHandler) removeTFA(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	removeTFARequest := struct {
		CurrentPassword string `json:"password" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &removeTFARequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(removeTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.RemoveTFA(user, removeTFARequest.CurrentPassword); err != nil {
		log.WithFields(log.Fields{
			"request":      removeTFARequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Change Email User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *ProfileHandler) deleteAccount(res http.ResponseWriter, req *http.Request) {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Email")

		RenderError(res, ErrSomethingWrong)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	removeTFARequest := struct {
		CurrentPassword string `json:"password" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &removeTFARequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(removeTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.DeleteAccount(user, removeTFARequest.CurrentPassword); err != nil {
		log.WithFields(log.Fields{
			"request":      removeTFARequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Change Email User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
