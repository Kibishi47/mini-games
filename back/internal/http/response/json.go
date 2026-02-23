package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func Success(w http.ResponseWriter, message string, status int) {
	JSON(w, status, map[string]string{
		"message": message,
	})
}

func Error(w http.ResponseWriter, message string, status int) {
	JSON(w, status, map[string]string{
		"error": message,
	})
}
