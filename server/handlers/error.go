package handlers

import (
	"errors"
	"net/http"

	"github.com/AdhityaRamadhanus/userland/server/render"
	"github.com/asaskevich/govalidator"
)

var (
	//ErrFailedToUnmarshalJSON request body processing error
	ErrFailedToUnmarshalJSON = errors.New("Failed to Unmarshal JSON")
	//ErrFailedToReadBody request body processing error
	ErrFailedToReadBody = errors.New("Failed to read request body")
	//ErrSomethingWrong request body processing error
	ErrSomethingWrong = errors.New("Something is wrong")
	//ErrInvalidRequest request body processing error
	ErrInvalidRequest = errors.New("Invalid Request Against this endpoint")
)

func RenderInvalidRequestError(res http.ResponseWriter, err error) error {
	fieldsError := govalidator.ErrorsByField(err)
	return render.JSON(res, http.StatusUnprocessableEntity, map[string]interface{}{
		"status": http.StatusUnprocessableEntity,
		"error": map[string]interface{}{
			"code":   "ErrInvalidRequest",
			"fields": fieldsError,
		},
	})
}

func RenderFailedToReadBodyError(res http.ResponseWriter, err error) error {
	return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrFailedToReadBody",
			"message": err.Error(),
		},
	})
}

func RenderFailedToUnmarshalJSONError(res http.ResponseWriter, err error) error {
	return render.JSON(res, http.StatusBadRequest, map[string]interface{}{
		"status": http.StatusBadRequest,
		"error": map[string]interface{}{
			"code":    "ErrFailedToUnmarshalJSON",
			"message": err.Error(),
		},
	})
}

func RenderInternalServerError(res http.ResponseWriter, err error) error {
	return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrInternalServer",
			"message": err.Error(),
		},
	})
}
