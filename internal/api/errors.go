package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode int
	Message    string
	Messages   []string
}

func (e *APIError) Error() string {
	if len(e.Messages) > 0 {
		return fmt.Sprintf("API error %d: %s", e.StatusCode, strings.Join(e.Messages, "; "))
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

func (e *APIError) IsValidation() bool {
	return e.StatusCode == 422
}

func parseError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}

	var singleErr struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &singleErr) == nil && singleErr.Error != "" {
		apiErr.Message = singleErr.Error
		return apiErr
	}

	var multiErr struct {
		Errors []string `json:"errors"`
	}
	if json.Unmarshal(body, &multiErr) == nil && len(multiErr.Errors) > 0 {
		apiErr.Messages = multiErr.Errors
		apiErr.Message = multiErr.Errors[0]
		return apiErr
	}

	apiErr.Message = string(body)
	return apiErr
}
