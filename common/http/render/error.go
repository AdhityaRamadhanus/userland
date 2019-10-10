package render

import (
	"net/http"

	"github.com/asaskevich/govalidator"
)

func InvalidRequestError(res http.ResponseWriter, err error) error {
	fieldsError := govalidator.ErrorsByField(err)
	return JSON(res, http.StatusUnprocessableEntity, map[string]interface{}{
		"status": http.StatusUnprocessableEntity,
		"error": map[string]interface{}{
			"code":   "ErrInvalidRequest",
			"fields": fieldsError,
		},
	})
}

func FailedToReadBodyError(res http.ResponseWriter, err error) error {
	return JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrFailedToReadBody",
			"message": err.Error(),
		},
	})
}

func FailedToUnmarshalJSONError(res http.ResponseWriter, err error) error {
	return JSON(res, http.StatusBadRequest, map[string]interface{}{
		"status": http.StatusBadRequest,
		"error": map[string]interface{}{
			"code":    "ErrFailedToUnmarshalJSON",
			"message": err.Error(),
		},
	})
}

func InternalServerError(res http.ResponseWriter, err error) error {
	return JSON(res, http.StatusInternalServerError, map[string]interface{}{
		"status": http.StatusInternalServerError,
		"error": map[string]interface{}{
			"code":    "ErrInternalServer",
			"message": err.Error(),
		},
	})
}
