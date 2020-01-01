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
	"time"

	"github.com/AdhityaRamadhanus/userland/pkg/server/api/serializers"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/asaskevich/govalidator"

	_ "image/jpeg"

	"github.com/AdhityaRamadhanus/userland/pkg/common/contextkey"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/render"
	"github.com/AdhityaRamadhanus/userland/pkg/service/event"
	"github.com/AdhityaRamadhanus/userland/pkg/service/profile"

	"github.com/gorilla/mux"
)

type ProfileHandler struct {
	Authorization  middlewares.MiddlewareWithArgs
	Authenticator  middlewares.Middleware
	RateLimiter    middlewares.MiddlewareWithArgs
	ProfileService profile.Service
	EventService   event.Service
}

func (h ProfileHandler) RegisterRoutes(router *mux.Router) {
	subRouter := router.PathPrefix("/api").Subrouter()

	authenticate := h.Authenticator
	authorize := h.Authorization
	ratelimit := h.RateLimiter

	getProfile := authenticate(authorize(http.HandlerFunc(h.getProfile), security.UserTokenScope))
	updateProfile := authenticate(authorize(http.HandlerFunc(h.updateProfile), security.UserTokenScope))
	setPicture := authenticate(authorize(http.HandlerFunc(h.setPicture), security.UserTokenScope))
	deletePicture := authenticate(authorize(http.HandlerFunc(h.deletePicture), security.UserTokenScope))
	getEmail := authenticate(authorize(http.HandlerFunc(h.getEmail), security.UserTokenScope))
	requestChangeEmail := ratelimit(authenticate(authorize(http.HandlerFunc(h.requestChangeEmail), security.UserTokenScope)), 10, time.Minute)
	changeEmail := authenticate(authorize(http.HandlerFunc(h.changeEmail), security.UserTokenScope))
	changePassword := ratelimit(authenticate(authorize(http.HandlerFunc(h.changePassword), security.UserTokenScope)), 10, time.Minute)
	getTFAStatus := authenticate(authorize(http.HandlerFunc(h.getTFAStatus), security.UserTokenScope))
	enrollTFA := authenticate(authorize(http.HandlerFunc(h.enrollTFA), security.UserTokenScope))
	activateTFA := authenticate(authorize(http.HandlerFunc(h.activateTFA), security.UserTokenScope))
	removeTFA := authenticate(authorize(http.HandlerFunc(h.removeTFA), security.UserTokenScope))
	deleteAccount := authenticate(authorize(http.HandlerFunc(h.deleteAccount), security.UserTokenScope))
	getEvents := authenticate(authorize(http.HandlerFunc(h.getEvents), security.UserTokenScope))

	subRouter.Handle("/me", getProfile).Methods("GET")
	subRouter.Handle("/me", updateProfile).Methods("POST")

	subRouter.Handle("/me/picture", setPicture).Methods("POST")
	subRouter.Handle("/me/picture", deletePicture).Methods("DELETE")

	subRouter.Handle("/me/email", getEmail).Methods("GET")
	subRouter.Handle("/me/email", requestChangeEmail).Methods("POST")
	subRouter.Handle("/me/email", changeEmail).Methods("PATCH")

	subRouter.Handle("/me/password", changePassword).Methods("POST")

	subRouter.Handle("/me/tfa", getTFAStatus).Methods("GET")
	subRouter.Handle("/me/tfa/enroll", enrollTFA).Methods("GET")
	subRouter.Handle("/me/tfa/enroll", activateTFA).Methods("POST")
	subRouter.Handle("/me/tfa/remove", removeTFA).Methods("POST")

	subRouter.Handle("/me/delete", deleteAccount).Methods("DELETE")
	subRouter.Handle("/me/events", getEvents).Methods("GET")
}

func (h ProfileHandler) getProfile(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, serializers.SerializeUserToJSON(user))
}

func (h ProfileHandler) updateProfile(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	updateProfileRequest := struct {
		Fullname string `json:"fullname" valid:"required,stringlength(3|128)"`
		Bio      string `json:"bio" valid:"optional,stringlength(1|255)"`
		Location string `json:"location" valid:"optional,stringlength(1|128)"`
		Web      string `json:"web" valid:"optional,stringlength(1|128)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &updateProfileRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(updateProfileRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
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
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) getEmail(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"email": user.Email})
}

func (h ProfileHandler) requestChangeEmail(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	userID := getUserIDFromContext(req)

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	changeEmailRequest := struct {
		Email string `json:"email" valid:"required,email,stringlength(1|128)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changeEmailRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(changeEmailRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	if _, err = h.ProfileService.RequestChangeEmail(user, changeEmailRequest.Email); err != nil {
		handleServiceError(res, req, err)
		return
	}

	defer h.EventService.Log(profile.EventChangeEmailRequest, userID, clientInfo)
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) changeEmail(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	userID := getUserIDFromContext(req)

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	changeEmailRequest := struct {
		Token string `json:"token" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changeEmailRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(changeEmailRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.ChangeEmail(user, changeEmailRequest.Token); err != nil {
		handleServiceError(res, req, err)
		return
	}

	defer h.EventService.Log(profile.EventChangeEmail, userID, clientInfo)
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) changePassword(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	userID := getUserIDFromContext(req)

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	changePasswordRequest := struct {
		CurrentPassword      string `json:"password_current" valid:"required,stringlength(6|128)"`
		NewPassword          string `json:"password" valid:"required,stringlength(6|128)"`
		NewConfirmedPassword string `json:"password_confirmed" valid:"required,stringlength(6|128)"`
		NewPasswordSame      string `valid:"required~Password should be same"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &changePasswordRequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if changePasswordRequest.NewPassword == changePasswordRequest.NewConfirmedPassword {
		changePasswordRequest.NewPasswordSame = "true"
	}

	if ok, err := govalidator.ValidateStruct(changePasswordRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	oldPassword := changePasswordRequest.CurrentPassword
	newPassword := changePasswordRequest.NewPassword
	if err = h.ProfileService.ChangePassword(user, oldPassword, newPassword); err != nil {
		handleServiceError(res, req, err)
		return
	}

	defer h.EventService.Log(profile.EventChangePassword, userID, clientInfo)
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) getTFAStatus(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	response := map[string]interface{}{"enabled_at": ""}
	if user.TFAEnabled {
		response["enabled"] = true
		response["enabled_at"] = user.TFAEnabledAt
	}
	render.JSON(res, http.StatusOK, response)
}

func (h ProfileHandler) enrollTFA(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	secret, qrCodeImageBase64, err := h.ProfileService.EnrollTFA(user)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"secret": secret,
		"qr":     fmt.Sprintf(`data:image/png;base64,%s`, qrCodeImageBase64),
	})
}

func (h ProfileHandler) activateTFA(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	userID := getUserIDFromContext(req)

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	activateTFARequest := struct {
		Secret string `json:"secret" valid:"required,stringlength(6|128)"`
		Code   string `json:"code" valid:"required,stringlength(6|6)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &activateTFARequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(activateTFARequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	backupCodes, err := h.ProfileService.ActivateTFA(user, activateTFARequest.Secret, activateTFARequest.Code)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	defer h.EventService.Log(profile.EventEnableTFA, userID, clientInfo)
	render.JSON(res, http.StatusOK, map[string]interface{}{"backup_codes": backupCodes})
}

func (h ProfileHandler) removeTFA(res http.ResponseWriter, req *http.Request) {
	clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
	userID := getUserIDFromContext(req)

	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	removeTFARequest := struct {
		CurrentPassword string `json:"password" valid:"required,stringlength(6|128)"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &removeTFARequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(removeTFARequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.RemoveTFA(user, removeTFARequest.CurrentPassword); err != nil {
		handleServiceError(res, req, err)
		return
	}

	defer h.EventService.Log(profile.EventDisableTFA, userID, clientInfo)
	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) deleteAccount(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		render.FailedToReadBodyError(res, err)
		return
	}

	removeTFARequest := struct {
		CurrentPassword string `json:"password" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &removeTFARequest); err != nil {
		render.FailedToUnmarshalJSONError(res, err)
		return
	}

	if err := req.Body.Close(); err != nil {
		render.InternalServerError(res, err)
		return
	}

	if ok, err := govalidator.ValidateStruct(removeTFARequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	if err = h.ProfileService.DeleteAccount(user, removeTFARequest.CurrentPassword); err != nil {
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) deletePicture(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	user.PictureURL = ""
	err = h.ProfileService.SetProfile(user)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) setPicture(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
	user, err := h.ProfileService.Profile(userID)
	if err != nil {
		handleServiceError(res, req, err)
		return
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			err = govalidator.Error{
				Name:      "File",
				Err:       err,
				Validator: "required",
				Path:      []string{"file"},
			}
			render.InvalidRequestError(res, err)
			return
		}
		handleServiceError(res, req, err)
		return
	}
	defer file.Close()

	// set max request body on load balancer/web server
	imageBytes, _ := ioutil.ReadAll(file)
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
	image, _, _ := image.DecodeConfig(bytes.NewReader(imageBytes))
	setPictureRequest.ImageWidth = image.Width
	setPictureRequest.ImageHeight = image.Height
	if image.Width == image.Height {
		setPictureRequest.ImageIsSquare = true
	}

	if ok, err := govalidator.ValidateStruct(setPictureRequest); !ok || err != nil {
		render.InvalidRequestError(res, err)
		return
	}

	if err := h.ProfileService.SetProfilePicture(user, bytes.NewReader(imageBytes)); err != nil {
		handleServiceError(res, req, err)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{"success": true})
}

func (h ProfileHandler) getEvents(res http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req)
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
		render.InvalidRequestError(res, err)
		return
	}

	filter := userland.EventFilterOptions{
		UserID: userID,
	}
	paging := userland.EventPagingOptions{
		Limit:  limit,
		Offset: (page - 1) * limit,
		SortBy: "timestamp",
		Order:  order,
	}
	events, eventsCount, err := h.EventService.ListEvents(filter, paging)
	if err != nil {
		handleServiceError(res, req, err)
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
		serializedEvent := serializers.SerializeEventToJSON(event)
		serializedEvents = append(serializedEvents, serializedEvent)
	}

	return serializedEvents
}
