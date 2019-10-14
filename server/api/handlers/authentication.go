package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/AdhityaRamadhanus/userland/server/api/serializers"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/common/http/render"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/service/authentication"
	"github.com/AdhityaRamadhanus/userland/service/event"
	"github.com/AdhityaRamadhanus/userland/service/profile"
	"github.com/AdhityaRamadhanus/userland/service/session"
	"github.com/asaskevich/govalidator"
	"github.com/go-errors/errors"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AuthenticationHandler struct {
	Authenticator         *middlewares.Authenticator
	RateLimiter           *middlewares.RateLimiter
	AuthenticationService authentication.Service
	SessionService        session.Service
	ProfileService        profile.Service
	EventService          event.Service
}

func (h AuthenticationHandler) RegisterRoutes(router *mux.Router) {
	authenticate := h.Authenticator.Authenticate
	tfaAuthorize := middlewares.Authorize(security.TFATokenScope)
	ratelimit3PerMinute := h.RateLimiter.Limit(10, time.Minute)

	router.HandleFunc("/auth/register", h.registerUser).Methods("POST")

	router.HandleFunc("/auth/verification", ratelimit3PerMinute(h.requestVerification)).Methods("POST")
	router.HandleFunc("/auth/verification", h.verifyAccount).Methods("PATCH")

	router.HandleFunc("/auth/login", h.login).Methods("POST")

	router.HandleFunc("/auth/password/forgot", ratelimit3PerMinute(h.forgotPassword)).Methods("POST")
	router.HandleFunc("/auth/password/reset", h.resetPassword).Methods("POST")

	router.HandleFunc("/auth/tfa/verify", authenticate(tfaAuthorize(h.verifyTFA))).Methods("POST")
	router.HandleFunc("/auth/tfa/bypass", authenticate(tfaAuthorize(h.verifyTFABypass))).Methods("POST")
}

func (h AuthenticationHandler) registerUser(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	registerUserRequest := struct {
		Fullname          string `json:"fullname" valid:"required,stringlength(1|128)"`
		Email             string `json:"email" valid:"required,email,stringlength(1|128)"`
		Password          string `json:"password" valid:"required,stringlength(6|128)"`
		ConfirmedPassword string `json:"password_confirmed" valid:"required,stringlength(6|128)"`
		PasswordSame      string `valid:"required~Password should be same"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &registerUserRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	//hack
	if registerUserRequest.Password == registerUserRequest.ConfirmedPassword {
		registerUserRequest.PasswordSame = "true"
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(registerUserRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	newUser := userland.User{
		Fullname: registerUserRequest.Fullname,
		Email:    registerUserRequest.Email,
		Password: registerUserRequest.Password,
	}

	if err = h.AuthenticationService.Register(newUser); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) requestVerification(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	requestVerificationRequest := struct {
		Type      string `json:"type" valid:"required,stringlength(1|32)"`
		Recipient string `json:"recipient" valid:"required,stringlength(1|128)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &requestVerificationRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(requestVerificationRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	verificationType := requestVerificationRequest.Type
	verificationRecipient := requestVerificationRequest.Recipient
	if _, err = h.AuthenticationService.RequestVerification(verificationType, verificationRecipient); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) verifyAccount(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	verifyAccountRequest := struct {
		Email          string `json:"email" valid:"required,email,stringlength(1|128)"`
		VerificationID string `json:"verification_id" valid:"required"`
		Code           string `json:"code" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyAccountRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyAccountRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	verificationID := verifyAccountRequest.VerificationID
	email := verifyAccountRequest.Email
	code := verifyAccountRequest.Code
	if err = h.AuthenticationService.VerifyAccount(verificationID, email, code); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) login(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	loginRequest := struct {
		Email    string `json:"email" valid:"required,email,stringlength(1|64)"`
		Password string `json:"password" valid:"required,stringlength(6|128)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &loginRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(loginRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	email := loginRequest.Email
	password := loginRequest.Password
	user, err := h.ProfileService.ProfileByEmail(email)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	requireTFA, accessToken, err := h.AuthenticationService.Login(email, password)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	if !requireTFA {
		h.SessionService.CreateSession(user.ID, userland.Session{
			ID:         accessToken.Key,
			Token:      accessToken.Value,
			IP:         clientInfo["ip"].(string),
			ClientID:   clientInfo["client_id"].(int),
			ClientName: clientInfo["client_name"].(string),
			Expiration: security.UserAccessTokenExpiration,
		})
	}

	defer h.EventService.Log(event.LoginEvent, user.ID, clientInfo)
	render.JSON(res, http.StatusOK, map[string]interface{}{
		"require_tfa":  requireTFA,
		"access_token": serializers.SerializeAccessTokenToJSON(accessToken),
	})
}

func (h AuthenticationHandler) forgotPassword(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	forgotPasswordRequest := struct {
		Email string `json:"email" valid:"required,email,stringlength(1|64)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &forgotPasswordRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(forgotPasswordRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	email := forgotPasswordRequest.Email
	if _, err = h.AuthenticationService.ForgotPassword(email); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	defer func(email string, clientInfo map[string]interface{}) {
		user, err := h.ProfileService.ProfileByEmail(email)
		if err == nil {
			h.EventService.Log(event.ForgotPasswordEvent, user.ID, clientInfo)
		}
	}(email, clientInfo)
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) resetPassword(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	resetPasswordRequest := struct {
		Token             string `json:"token" valid:"required"`
		Password          string `json:"password" valid:"required,stringlength(6|128)"`
		ConfirmedPassword string `json:"password_confirmed" valid:"required,stringlength(6|128)"`
		PasswordSame      string `valid:"required~Password should be same"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &resetPasswordRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if resetPasswordRequest.Password == resetPasswordRequest.ConfirmedPassword {
		resetPasswordRequest.PasswordSame = "true"
	}

	if ok, err := govalidator.ValidateStruct(resetPasswordRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	resetToken := resetPasswordRequest.Token
	newPassword := resetPasswordRequest.Password
	if err = h.AuthenticationService.ResetPassword(resetToken, newPassword); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) verifyTFA(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	tfaAccessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	tfaAccessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(tfaAccessToken["userid"].(float64))

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	verifyTFARequest := struct {
		Code string `json:"code" valid:"required,stringlength(6|6)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyTFARequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyTFARequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	accessToken, err := h.AuthenticationService.VerifyTFA(tfaAccessTokenKey, userID, verifyTFARequest.Code)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// suppress error
	h.SessionService.CreateSession(userID, userland.Session{
		ID:         accessToken.Key,
		Token:      accessToken.Value,
		IP:         clientInfo["ip"].(string),
		ClientID:   clientInfo["client_id"].(int),
		ClientName: clientInfo["client_name"].(string),
		Expiration: security.UserAccessTokenExpiration,
	})

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"access_token": serializers.SerializeAccessTokenToJSON(accessToken),
	})
}

func (h AuthenticationHandler) verifyTFABypass(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	tfaAccessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	tfaAccessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(tfaAccessToken["userid"].(float64))
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	verifyTFARequest := struct {
		Code string `json:"code" valid:"required,stringlength(6|6)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyTFARequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyTFARequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	accessToken, err := h.AuthenticationService.VerifyTFABypass(tfaAccessTokenKey, userID, verifyTFARequest.Code)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	h.SessionService.CreateSession(userID, userland.Session{
		ID:         accessToken.Key,
		Token:      accessToken.Value,
		IP:         clientInfo["ip"].(string),
		ClientID:   clientInfo["client_id"].(int),
		ClientName: clientInfo["client_name"].(string),
		Expiration: security.UserAccessTokenExpiration,
	})

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"access_token": serializers.SerializeAccessTokenToJSON(accessToken),
	})
}

func (h AuthenticationHandler) handleServiceError(res http.ResponseWriter, req *http.Request, err error) {
	ServiceErrorsHTTPMapping := map[error]struct {
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
		"stack_trace":  fmt.Sprintf("%v", err.(*errors.Error).ErrorStack()),
		"endpoint":     req.URL.Path,
		"client":       req.Header.Get("X-API-ClientID"),
		"x-request-id": req.Header.Get("X-Request-ID"),
	}).WithError(err).Error("Error Authentication Handler")

	render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrInternalServer",
			"message": "userland api unable to process the request",
		},
	})
	return
}
