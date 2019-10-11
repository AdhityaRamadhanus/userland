package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/common/http/render"
	"github.com/AdhityaRamadhanus/userland/service/mailing"
	"github.com/asaskevich/govalidator"
	"github.com/go-errors/errors"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type MailHandler struct {
	MailingService mailing.Service
}

func (h MailHandler) RegisterRoutes(router *mux.Router) {
	basicAuth := middlewares.BasicAuth(os.Getenv("MAIL_SERVICE_BASIC_USER"), os.Getenv("MAIL_SERVICE_BASIC_PASS"))
	router.HandleFunc("/mail/otp", basicAuth(h.sendEmailOTP)).Methods("POST")
	router.HandleFunc("/mail/verification", basicAuth(h.sendEmailVerification)).Methods("POST")
}

func (h *MailHandler) sendEmailOTP(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	sendEmailOTPRequest := struct {
		OTP           string `json:"otp" valid:"required,stringlength(6|6)"`
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
		h.handleServiceError(res, req, err)
		return
	}
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *MailHandler) sendEmailVerification(res http.ResponseWriter, req *http.Request) {
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
		h.handleServiceError(res, req, err)
		return
	}
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h MailHandler) handleServiceError(res http.ResponseWriter, req *http.Request, err error) {
	log.WithFields(log.Fields{
		"stack_trace":  fmt.Sprintf("%v", err.(*errors.Error).ErrorStack()),
		"endpoint":     req.URL.Path,
		"x-request-id": req.Header.Get("X-Request-ID"),
	}).WithError(err).Error("Error Mail Handler")

	render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrInternalServer",
			"message": "userland mail unable to process the request",
		},
	})
	return
}
