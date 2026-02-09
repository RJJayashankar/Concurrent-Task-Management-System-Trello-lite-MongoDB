package utils

import (
	"encoding/json"
	"net/http"
)

// Response structure matching your request
type APIResponse struct {
	Status string      `json:"status"`
	Desc   string      `json:"desc"`
	Data   interface{} `json:"data,omitempty"`
}

// SendSuccess sends a standardized 200 OK response
func SendSuccess(w http.ResponseWriter, desc string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := APIResponse{
		Status: "Success",
		Desc:   desc,
		Data:   data,
	}
	json.NewEncoder(w).Encode(response)
}

// SendError sends a standardized error response
func SendError(w http.ResponseWriter, code int, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := APIResponse{
		Status: "Error",
		Desc:   desc,
	}
	json.NewEncoder(w).Encode(response)
}
