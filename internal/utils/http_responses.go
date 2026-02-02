package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, ErrorResponse{
		Title:   "Error",
		Message: message,
	})
}

func RespondSuccess(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, SuccessResponse{
		Title:   "Success",
		Message: message,
	})
}
