package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/asaskevich/govalidator"

	_ "image/jpeg"

	"github.com/AdhityaRamadhanus/userland/profile"
	"github.com/AdhityaRamadhanus/userland/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/userland/server/middlewares"
	"github.com/AdhityaRamadhanus/userland/server/render"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type ProfileHandler struct {
	Authenticator        *middlewares.Authenticator
	ProfileService       profile.Service
	ObjectStorageService userland.ObjectStorageService
}

func (h ProfileHandler) RegisterRoutes(router *mux.Router) {
	authenticate := h.Authenticator.Authenticate
	authorize := middlewares.Authorize
	router.HandleFunc("/me", authenticate(authorize(security.UserTokenScope, h.getProfile))).Methods("GET")
	router.HandleFunc("/me", authenticate(authorize(security.UserTokenScope, h.updateProfile))).Methods("POST")

	router.HandleFunc("/me/picture", authenticate(authorize(security.UserTokenScope, h.setPicture))).Methods("POST")
	router.HandleFunc("/me/picture", authenticate(authorize(security.UserTokenScope, h.deletePicture))).Methods("DELETE")

	router.HandleFunc("/me/email", authenticate(authorize(security.UserTokenScope, h.getEmail))).Methods("GET")
	router.HandleFunc("/me/email", authenticate(authorize(security.UserTokenScope, h.requestChangeEmail))).Methods("POST")
	router.HandleFunc("/me/email", authenticate(authorize(security.UserTokenScope, h.changeEmail))).Methods("PATCH")

	router.HandleFunc("/me/password", authenticate(authorize(security.UserTokenScope, h.changePassword))).Methods("POST")

	router.HandleFunc("/me/tfa", authenticate(authorize(security.UserTokenScope, h.getTFAStatus))).Methods("GET")
	router.HandleFunc("/me/tfa/enroll", authenticate(authorize(security.UserTokenScope, h.enrollTFA))).Methods("GET")
	router.HandleFunc("/me/tfa/enroll", authenticate(authorize(security.UserTokenScope, h.activateTFA))).Methods("POST")
	router.HandleFunc("/me/tfa/remove", authenticate(authorize(security.UserTokenScope, h.removeTFA))).Methods("POST")

	router.HandleFunc("/me/delete", authenticate(authorize(security.UserTokenScope, h.deleteAccount))).Methods("POST")

	router.HandleFunc("/me/events", authenticate(authorize(security.UserTokenScope, h.getEvents))).Methods("GET")
}

func (h *ProfileHandler) getProfile(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
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
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
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
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
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
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) getEmail(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"email": user.Email})
}

func (h *ProfileHandler) requestChangeEmail(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	changeEmailRequest := struct {
		Email string `json:"email" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changeEmailRequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(changeEmailRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if _, err = h.ProfileService.RequestChangeEmail(user, changeEmailRequest.Email); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) changeEmail(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	changeEmailRequest := struct {
		Token string `json:"token" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changeEmailRequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(changeEmailRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.ChangeEmail(user, changeEmailRequest.Token); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) changePassword(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	changePasswordRequest := struct {
		CurrentPassword      string `json:"password_current" valid:"required"`
		NewPassword          string `json:"password" valid:"required"`
		NewConfirmedPassword string `json:"password_confirmed" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changePasswordRequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(changePasswordRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	oldPassword := changePasswordRequest.CurrentPassword
	newPassword := changePasswordRequest.NewPassword
	if err = h.ProfileService.ChangePassword(user, oldPassword, newPassword); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) getTFAStatus(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	response := map[string]interface{}{"enabled_at": ""}
	if user.TFAEnabled {
		response["enabled"] = true
		response["enabled_at"] = user.TFAEnabledAt
	}
	render.JSON(res, http.StatusOK, response)
}

func (h *ProfileHandler) enrollTFA(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	secret, qrCodeImageBase64, err := h.ProfileService.EnrollTFA(user)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"secret": secret,
		"qr":     fmt.Sprintf(`data:image/png;base64,%s`, qrCodeImageBase64),
	})
}

func (h *ProfileHandler) activateTFA(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	activateTFARequest := struct {
		Secret string `json:"secret" valid:"required"`
		Code   string `json:"code" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &activateTFARequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(activateTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	backupCodes, err := h.ProfileService.ActivateTFA(user, activateTFARequest.Secret, activateTFARequest.Code)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"backup_codes": backupCodes})
}

func (h *ProfileHandler) removeTFA(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	removeTFARequest := struct {
		CurrentPassword string `json:"password" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &removeTFARequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(removeTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.RemoveTFA(user, removeTFARequest.CurrentPassword); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) deleteAccount(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderFailedToReadBodyError(res, err)
		return
	}

	removeTFARequest := struct {
		CurrentPassword string `json:"password" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &removeTFARequest); err != nil {
		RenderFailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderInternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(removeTFARequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.DeleteAccount(user, removeTFARequest.CurrentPassword); err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) deletePicture(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	user.PictureURL = ""
	err = h.ProfileService.SetProfile(user)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) setPicture(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}
	defer file.Close()
	var imageBuffer bytes.Buffer
	duplicateFile := io.TeeReader(file, &imageBuffer)
	imageBufferReader := bytes.NewReader(imageBuffer.Bytes())

	// validate image
	setPictureRequest := struct {
		FileName      string `valid:"required"`
		FileSize      int    `valid:"required"`
		ImageWidth    int    `valid:"required"`
		ImageHeight   int    `valid:"required"`
		ImageIsSquare bool   `valid:"required~Image is not square"`
	}{}
	setPictureRequest.FileName = header.Filename
	setPictureRequest.FileSize = int(header.Size)
	image, _, _ := image.DecodeConfig(duplicateFile)
	setPictureRequest.ImageWidth = image.Width
	setPictureRequest.ImageHeight = image.Height
	if image.Width == image.Height {
		setPictureRequest.ImageIsSquare = true
	}

	if ok, err := govalidator.ValidateStruct(setPictureRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	link, err := h.ObjectStorageService.Write(imageBufferReader, userland.ObjectMetaData{
		CacheControl: "public, max-age=86400",
		ContentType:  "image/jpeg",
		Path:         fmt.Sprintf("userland_%s_profile.jpeg", user.Fullname),
	})
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	user.PictureURL = link
	err = h.ProfileService.SetProfile(user)
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *ProfileHandler) getEvents(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	page, _ := strconv.Atoi(req.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	order := req.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	// filter
	getEventsRequest := struct {
		Limit int    `valid:"int"`
		Page  int    `valid:"int"`
		Order string `valid:"in(asc|desc)"`
	}{
		Limit: limit,
		Page:  page,
		Order: order,
	}

	if ok, err := govalidator.ValidateStruct(getEventsRequest); !ok || err != nil {
		RenderInvalidRequestError(res, err)
		return
	}

	events, eventsCount, err := h.ProfileService.ListEvents(user, userland.EventPagingOptions{
		Limit:  limit,
		Offset: (page - 1) * limit,
		SortBy: "timestamp",
		Order:  order,
	})
	if err != nil {
		h.handleServiceError(res, req, err)
		return
	}

	totalPage := int(math.Ceil(float64(eventsCount) / float64(limit)))
	nextPage := page + 1
	if nextPage > totalPage {
		nextPage = totalPage
	}
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"data":   serializeEvents(events),
		"pagination": map[string]interface{}{
			"next":     nextPage,
			"previous": prevPage,
			"per_page": limit,
			"page":     totalPage,
		},
	})
}

func getUserIDFromContext(req *http.Request) int {
	accessToken := req.Context().Value(contextkey.AccessToken).(map[string]interface{})
	userID := int(accessToken["userid"].(float64))
	return userID
}

func serializeEvents(events []userland.Event) []map[string]interface{} {
	serializedEvents := []map[string]interface{}{}
	for _, event := range events {
		serializedEvent := map[string]interface{}{
			"event": event.Event,
			"ua":    event.UserAgent,
			"ip":    event.IP,
			"client": map[string]interface{}{
				"id":   event.ClientID,
				"name": event.ClientName,
			},
			"created_at": event.Timestamp,
		}

		serializedEvents = append(serializedEvents, serializedEvent)
	}

	return serializedEvents
}

func (h *ProfileHandler) handleServiceError(res http.ResponseWriter, req *http.Request, err error) {
	ServiceErrorsHTTPMapping := map[error]struct {
		HTTPCode int
		ErrCode  string
	}{
		userland.ErrUserNotFound: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrUserNotFound",
		},
		profile.ErrWrongOTP: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrWrongOTP",
		},
		profile.ErrWrongPassword: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrWrongPassword",
		},
		profile.ErrEmailAlreadyUsed: {
			HTTPCode: http.StatusNotFound,
			ErrCode:  "ErrEmailAlreadyUsed",
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
	}).WithError(err).Error("Error Profile Handler")

	render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrInternalServer",
			"message": "userland api unable to process the request",
		},
	})
	return
}
