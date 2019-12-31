package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
	"github.com/AdhityaRamadhanus/userland/pkg/service/mailing"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

type MailingHandler struct {
	Authenticator  middlewares.Middleware
	MailingService mailing.Service
}

func (h MailingHandler) RegisterRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/api").Subrouter()

	authenticate := h.Authenticator

	sendEmailOTP := authenticate(http.HandlerFunc(h.sendEmailOTP))
	sendEmailVerification := authenticate(http.HandlerFunc(h.sendEmailVerification))

	subRouter.Handle("/mail/otp", sendEmailOTP).Methods("POST")
	subRouter.Handle("/mail/verification", sendEmailVerification).Methods("POST")
}

func (h MailingHandler) sendEmailOTP(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	sendEmailOTPRequest := struct {
		OTP           string `json:"otp" valid:"required,stringlength(6|50)"`
		Type          string `json:"type" valid:"required,stringlength(1|128)"`
		Recipient     string `json:"recipient" valid:"required,email,stringlength(6|128)"`
		RecipientName string `json:"recipient_name" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &sendEmailOTPRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(sendEmailOTPRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	recipient := mailing.MailAddress{
		Address: sendEmailOTPRequest.Recipient,
		Name:    sendEmailOTPRequest.RecipientName,
	}
	otpType := sendEmailOTPRequest.Type
	otp := sendEmailOTPRequest.OTP

	if err := h.MailingService.SendOTPEmail(recipient, otpType, otp); err != nil {
		handleServiceError(res, req, err)
		return
	}
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h MailingHandler) sendEmailVerification(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	verificationEmailRequest := struct {
		VerificationLink string `json:"verification_link" valid:"required,stringlength(1|128)"`
		Recipient        string `json:"recipient" valid:"required,email,stringlength(6|128)"`
		RecipientName    string `json:"recipient_name" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &verificationEmailRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(verificationEmailRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	recipient := mailing.MailAddress{
		Address: verificationEmailRequest.Recipient,
		Name:    verificationEmailRequest.RecipientName,
	}
	verificationLink := verificationEmailRequest.VerificationLink

	if err := h.MailingService.SendVerificationEmail(recipient, verificationLink); err != nil {
		handleServiceError(res, req, err)
		return
	}
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}
