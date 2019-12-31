package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/AdhityaRamadhanus/userland/pkg/server/api/serializers"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/service/authentication"
	"github.com/AdhityaRamadhanus/userland/pkg/service/event"
	"github.com/AdhityaRamadhanus/userland/pkg/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/service/session"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

type AuthenticationHandler struct {
	Authenticator         middlewares.Middleware
	RateLimiter           middlewares.MiddlewareWithArgs
	Authorization         middlewares.MiddlewareWithArgs
	AuthenticationService authentication.Service
	SessionService        session.Service
	ProfileService        profile.Service
	EventService          event.Service
}

func (h AuthenticationHandler) RegisterRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/api").Subrouter()
	// middlewares
	authenticate := h.Authenticator
	authorize := h.Authorization
	ratelimit := h.RateLimiter

	registerUser := http.HandlerFunc(h.registerUser)
	requestVerification := ratelimit(http.HandlerFunc(h.requestVerification), 10, time.Minute)
	verifyAccount := http.HandlerFunc(h.verifyAccount)
	login := http.HandlerFunc(h.login)
	forgotPassword := ratelimit(http.HandlerFunc(h.forgotPassword), 10, time.Minute)
	resetPassword := http.HandlerFunc(h.resetPassword)
	verifyTFA := authenticate(authorize(http.HandlerFunc(h.verifyTFA), security.TFATokenScope))
	verifyTFABypass := authenticate(authorize(http.HandlerFunc(h.verifyTFABypass), security.TFATokenScope))

	subRouter.Handle("/auth/register", registerUser).Methods("POST")

	subRouter.Handle("/auth/verification", requestVerification).Methods("POST")
	subRouter.Handle("/auth/verification", verifyAccount).Methods("PATCH")

	subRouter.Handle("/auth/login", login).Methods("POST")

	subRouter.Handle("/auth/password/forgot", forgotPassword).Methods("POST")
	subRouter.Handle("/auth/password/reset", resetPassword).Methods("POST")

	subRouter.Handle("/auth/tfa/verify", verifyTFA).Methods("POST")
	subRouter.Handle("/auth/tfa/bypass", verifyTFABypass).Methods("POST")
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
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
		return
	}

	requireTFA, accessToken, err := h.AuthenticationService.Login(email, password)
	if err != nil {
		handleServiceError(res, req, err)
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

	defer h.EventService.Log(authentication.EventLogin, user.ID, clientInfo)
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
		handleServiceError(res, req, err)
		return
	}

	defer func(email string, clientInfo map[string]interface{}) {
		user, err := h.ProfileService.ProfileByEmail(email)
		if err == nil {
			h.EventService.Log(authentication.EventForgotPassword, user.ID, clientInfo)
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
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
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
		handleServiceError(res, req, err)
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
