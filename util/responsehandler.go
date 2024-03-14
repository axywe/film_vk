package util

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func SendJSONError(w http.ResponseWriter, r *http.Request, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorResponse := ErrorResponse{
		Code:    code,
		Message: message,
	}
	log.Printf(`%d | %s | %s "%s" | %s`, code, r.RemoteAddr, r.Method, r.RequestURI, message)
	err := json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func SendJSONResponse(w http.ResponseWriter, r *http.Request, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	log.Printf(`%d | %s | %s "%s"`, code, r.RemoteAddr, r.Method, r.RequestURI)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
