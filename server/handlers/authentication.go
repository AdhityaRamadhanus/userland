package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/authentication"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/userland/server/middlewares"
	"github.com/AdhityaRamadhanus/userland/server/render"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AuthenticationHandler struct {
	Authenticator         *middlewares.Authenticator
	AuthenticationService authentication.Service
}

func (h AuthenticationHandler) RegisterRoutes(router *mux.Router) {
	authenticate := h.Authenticator.Authenticate
	authorize := middlewares.Authorize
	router.HandleFunc("/auth/register", h.registerUser).Methods("POST")
	router.HandleFunc("/auth/verification", h.requestVerification).Methods("POST")
	router.HandleFunc("/auth/verification", h.verifyAccount).Methods("PATCH")
	router.HandleFunc("/auth/login", h.login).Methods("POST")
	router.HandleFunc("/auth/password/forgot", h.forgotPassword).Methods("POST")
	router.HandleFunc("/auth/password/reset", h.resetPassword).Methods("POST")
	router.HandleFunc("/auth/tfa/verify", authenticate(authorize(security.TFATokenScope, h.verifyTFA))).Methods("POST")
	router.HandleFunc("/auth/tfa/bypass", authenticate(authorize(security.TFATokenScope, h.verifyTFABypass))).Methods("POST")
}

func (h *AuthenticationHandler) registerUser(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	registerUserRequest := struct {
		Fullname          string `json:"fullname" valid:"required"`
		Email             string `json:"email" valid:"required"`
		Password          string `json:"password" valid:"required"`
		ConfirmedPassword string `json:"password_confirmed" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &registerUserRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(registerUserRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	newUser := userland.User{
		Fullname: registerUserRequest.Fullname,
		Email:    registerUserRequest.Email,
		Password: registerUserRequest.Password,
	}

	if err = h.AuthenticationService.Register(newUser); err != nil {
		log.WithFields(log.Fields{
			"request":      registerUserRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Register User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"success": true,
	})
}

func (h *AuthenticationHandler) requestVerification(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	requestVerificationRequest := struct {
		Type      string `json:"type" valid:"required"`
		Recipient string `json:"recipient" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &requestVerificationRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(requestVerificationRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if _, err = h.AuthenticationService.RequestVerification(requestVerificationRequest.Type, requestVerificationRequest.Recipient); err != nil {
		log.WithFields(log.Fields{
			"request":      requestVerificationRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Request Verification User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"success": true,
	})
}

func (h *AuthenticationHandler) verifyAccount(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	verifyAccountRequest := struct {
		Email          string `json:"email" valid:"required"`
		VerificationID string `json:"verification_id" valid:"required"`
		Code           string `json:"code" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyAccountRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyAccountRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	verificationID := verifyAccountRequest.VerificationID
	email := verifyAccountRequest.Email
	code := verifyAccountRequest.Code
	if err = h.AuthenticationService.VerifyAccount(verificationID, email, code); err != nil {
		log.WithFields(log.Fields{
			"request":      verifyAccountRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Verify Account User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"success": true,
	})
}

func (h *AuthenticationHandler) login(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	loginRequest := struct {
		Email    string `json:"email" valid:"required"`
		Password string `json:"password" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &loginRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(loginRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	email := loginRequest.Email
	password := loginRequest.Password
	requireTFA, accessToken, err := h.AuthenticationService.Login(email, password)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      loginRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Login User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"require_tfa": requireTFA,
		"access_token": map[string]interface{}{
			"value":      accessToken.Key,
			"type":       accessToken.Type,
			"expired_at": accessToken.ExpiredAt,
		},
	})
}

func (h *AuthenticationHandler) forgotPassword(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	forgotPasswordRequest := struct {
		Email string `json:"email" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &forgotPasswordRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(forgotPasswordRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	email := forgotPasswordRequest.Email
	_, err = h.AuthenticationService.ForgotPassword(email)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      forgotPasswordRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Forgot Password User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"success": true,
	})
}

func (h *AuthenticationHandler) resetPassword(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	resetPasswordRequest := struct {
		Token             string `json:"token" valid:"required"`
		Password          string `json:"password" valid:"required"`
		ConfirmedPassword string `json:"password_confirmed" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &resetPasswordRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(resetPasswordRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	err = h.AuthenticationService.ResetPassword(resetPasswordRequest.Token, resetPasswordRequest.Password)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      resetPasswordRequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Reset Password User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"success": true,
	})
}

func (h *AuthenticationHandler) verifyTFA(res http.ResponseWriter, req *http.Request) {
	tfaAccessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	tfaAccessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(tfaAccessToken["userid"].(float64))
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	verifyTFARequest := struct {
		Code string `json:"code" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyTFARequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	accessToken, err := h.AuthenticationService.VerifyTFA(tfaAccessTokenKey, userID, verifyTFARequest.Code)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      verifyTFARequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Verify TFA User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"access_token": map[string]interface{}{
			"value":      accessToken.Key,
			"type":       accessToken.Type,
			"expired_at": accessToken.ExpiredAt,
		},
	})
}

func (h *AuthenticationHandler) verifyTFABypass(res http.ResponseWriter, req *http.Request) {
	tfaAccessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	tfaAccessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(tfaAccessToken["userid"].(float64))
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	verifyTFARequest := struct {
		Code string `json:"code" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyTFARequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	accessToken, err := h.AuthenticationService.VerifyTFABypass(tfaAccessTokenKey, userID, verifyTFARequest.Code)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      verifyTFARequest,
			"client":       req.Header.Get("X-API-ClientID"),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Verify TFA Bypass User")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"access_token": map[string]interface{}{
			"value":      accessToken.Key,
			"type":       accessToken.Type,
			"expired_at": accessToken.ExpiredAt,
		},
	})
}
