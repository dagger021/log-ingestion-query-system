package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
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

func parseArray(values url.Values, key string) []string {
	if val := values[key]; len(val) > 0 {
		return val
	}
	return nil
}

func parseUint(values url.Values, key string) uint64 {
	if val := values.Get(key); val != "" {
		if val, err := strconv.ParseUint(val, 10, 64); err == nil {
			return val
		}
	}
	return 0
}

func parseOne(values url.Values, key string) *string {
	if val := values.Get(key); val != "" {
		return &val
	}
	return nil
}

func parseTime(values url.Values, key string) *time.Time {
	if val := values.Get("from"); val != "" {
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return &t
		}
	}
	return nil
}
