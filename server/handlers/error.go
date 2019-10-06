package handlers

import (
	"errors"
	"net/http"
	"strings"

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

//RenderError help handler create a consistent error response
func RenderError(res http.ResponseWriter, err error, customMessages ...string) error {
	errorMessage := err.Error()
	if len(customMessages) > 0 {
		errorMessage = strings.Join(customMessages, " ")
	}

	switch err {
	case ErrFailedToReadBody:
		return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
			"status": http.StatusInternalServerError,
			"error": map[string]interface{}{
				"code":    "ErrFailedToReadBody",
				"message": errorMessage,
			},
		})
	case ErrFailedToUnmarshalJSON:
		return render.JSON(res, http.StatusBadRequest, map[string]interface{}{
			"status": http.StatusBadRequest,
			"error": map[string]interface{}{
				"code":    "ErrFailedToUnmarshalJSON",
				"message": errorMessage,
			},
		})
	case ErrSomethingWrong:
		return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
			"status": http.StatusInternalServerError,
			"error": map[string]interface{}{
				"code":    "ErrInternalServer",
				"message": errorMessage,
			},
		})
	default:
		return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
			"status": http.StatusInternalServerError,
			"error": map[string]interface{}{
				"code":    "ErrInternalServer",
				"message": errorMessage,
			},
		})
	}
}
