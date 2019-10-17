package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

func CreateJSONRequest(method string, path string, requestBody interface{}) (*http.Request, error) {
	var httpReq *http.Request
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		jsonReqBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, errors.New("Failed to marshal request body")
		}
		httpReq, err = http.NewRequest(method, path, bytes.NewBuffer(jsonReqBody))
		if err != nil {
			return nil, errors.New("Failed to create request body")
		}
		httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	case http.MethodGet:
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			return nil, errors.New("Failed to marshal get request body")
		}
		httpReq = req
	}

	return httpReq, nil
}
