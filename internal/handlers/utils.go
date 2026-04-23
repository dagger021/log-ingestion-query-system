package handlers

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Err string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func WriteError(w http.ResponseWriter, status int, err string) {
	WriteJSON(w, status, errorResponse{Err: err})
}
