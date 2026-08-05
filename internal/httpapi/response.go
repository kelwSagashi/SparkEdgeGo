package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
)

type HandlerFunc func(r *http.Request) (any, error)

type ResponsePayload struct {
	Status  int
	Headers map[string]string
	Body    any
}

type HTTPError struct {
	Status  int
	Code    int
	Message string
	Hint    string
	Meta    map[string]any
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(status int, message string) *HTTPError {
	return &HTTPError{
		Status:  status,
		Code:    0,
		Message: message,
	}
}

func Adapt(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := handler(r)
		if err != nil {
			RespondError(w, err)
			return
		}
		Respond(w, http.StatusOK, data)
	}
}

func Respond(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")

	if response, ok := payload.(ResponsePayload); ok {
		for key, value := range response.Headers {
			w.Header().Add(key, value)
		}
		if response.Status != 0 {
			status = response.Status
		}
		payload = response.Body
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func RespondError(w http.ResponseWriter, err error) {
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		httpErr = NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	response := map[string]any{
		"code":    httpErr.Code,
		"message": httpErr.Message,
	}

	if httpErr.Hint != "" {
		response["hint"] = httpErr.Hint
	}
	if httpErr.Meta != nil {
		response["meta"] = httpErr.Meta
	}

	Respond(w, httpErr.Status, response)
}
