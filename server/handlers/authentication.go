package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/authentication"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/profile"
	"github.com/AdhityaRamadhanus/userland/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/userland/server/middlewares"
	"github.com/AdhityaRamadhanus/userland/server/render"
	"github.com/AdhityaRamadhanus/userland/session"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AuthenticationHandler struct {
	Authenticator         *middlewares.Authenticator
	AuthenticationService authentication.Service
	SessionService        session.Service
	ProfileService        profile.Service
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

func (h AuthenticationHandler) registerUser(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
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
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	//hack
	if registerUserRequest.Password == registerUserRequest.ConfirmedPassword {
		registerUserRequest.PasswordSame = "true"
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
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
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) requestVerification(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	requestVerificationRequest := struct {
		Type      string `json:"type" valid:"required,stringlength(1|32)"`
		Recipient string `json:"recipient" valid:"required,stringlength(1|128)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &requestVerificationRequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(requestVerificationRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if _, err = h.AuthenticationService.RequestVerification(requestVerificationRequest.Type, requestVerificationRequest.Recipient); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) verifyAccount(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	verifyAccountRequest := struct {
		Email          string `json:"email" valid:"required,email,stringlength(1|128)"`
		VerificationID string `json:"verification_id" valid:"required"`
		Code           string `json:"code" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyAccountRequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
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
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) login(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	loginRequest := struct {
		Email    string `json:"email" valid:"required,email,stringlength(1|64)"`
		Password string `json:"password" valid:"required,stringlength(6|128)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &loginRequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
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
		h.handleServiceError(res, req, err)
		return
	}

	if !requireTFA {
		user, err := h.ProfileService.ProfileByEmail(email)
		if err != nil {
			h.handleServiceError(res, req, err)
			return
		}
		h.SessionService.CreateSession(user.ID, userland.Session{
			ID:         accessToken.Key,
			Token:      accessToken.Value,
			IP:         clientInfo["ip"].(string),
			ClientID:   clientInfo["client_id"].(int),
			ClientName: clientInfo["client_name"].(string),
		})
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

func (h AuthenticationHandler) forgotPassword(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	forgotPasswordRequest := struct {
		Email string `json:"email" valid:"required,email,stringlength(1|64)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &forgotPasswordRequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(forgotPasswordRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	email := forgotPasswordRequest.Email
	_, err = h.AuthenticationService.ForgotPassword(email)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) resetPassword(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
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
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if resetPasswordRequest.Password == resetPasswordRequest.ConfirmedPassword {
		resetPasswordRequest.PasswordSame = "true"
	}

	if ok, err := govalidator.ValidateStruct(resetPasswordRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	err = h.AuthenticationService.ResetPassword(resetPasswordRequest.Token, resetPasswordRequest.Password)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{"success": true})
}

func (h AuthenticationHandler) verifyTFA(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	tfaAccessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	tfaAccessTokenKey := req.Context().Value(contextkey.AccessTokenKey).(string)
	userID := int(tfaAccessToken["userid"].(float64))

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	verifyTFARequest := struct {
		Code string `json:"code" valid:"required,stringlength(6|6)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyTFARequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
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
	})

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"access_token": map[string]interface{}{
			"value":      accessToken.Key,
			"type":       accessToken.Type,
			"expired_at": accessToken.ExpiredAt,
		},
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
		RenderFailedToReadBodyError(res, err)
		return
	}

	verifyTFARequest := struct {
		Code string `json:"code" valid:"required,stringlength(6|6)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verifyTFARequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(verifyTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
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
	})

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"access_token": map[string]interface{}{
			"value":      accessToken.Key,
			"type":       accessToken.Type,
			"expired_at": accessToken.ExpiredAt,
		},
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
