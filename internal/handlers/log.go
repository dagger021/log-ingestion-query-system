package handlers

import (
	"net/http"

	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/internal/services"
)

type LogEntryHandler interface {
	GetLogs(http.ResponseWriter, *http.Request)

	PostLog(http.ResponseWriter, *http.Request)
}

type logEntryHandler struct {
	svc services.LogEntryService
}

// GetLogs implements [LogEntryHandler].
func (h *logEntryHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	logEntries, err := h.svc.GetLogs(r.Context())
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	WriteJSON(w, http.StatusOK, logEntries)
}

// PostLog implements [LogEntryHandler].
func (h *logEntryHandler) PostLog(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if n, err := r.Body.Read(body); err != nil || n == 0 {
		WriteError(w, http.StatusBadRequest, "no body")
		return
	}

	logEntry, err := domain.ParseLogFromJSON(body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.StoreLog(r.Context(), *logEntry); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, "log processed")
}

func NewLogEntryHandler(svc services.LogEntryService) LogEntryHandler {
	return &logEntryHandler{svc: svc}
}
