package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/internal/logger"
	"github.com/dagger021/log-ingestion-query-system/internal/services"
	"go.uber.org/zap"
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
	ctx := r.Context()
	log := logger.FromContext(ctx)

	logEntries, err := h.svc.GetLogs(ctx)
	if err != nil {
		log.Error("error selecting logEntries", zap.Error(err))
		WriteError(w, http.StatusInternalServerError, "invalid response payload")
		return
	}

	WriteJSON(w, http.StatusOK, logEntries)
}

// PostLog implements [LogEntryHandler].
func (h *logEntryHandler) PostLog(w http.ResponseWriter, r *http.Request) {
	var logEntry domain.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&logEntry); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()
	log := logger.FromContext(ctx)
	if err := h.svc.StoreLog(ctx, logEntry); err != nil {
		log.Error("error storing logEntry", zap.Error(err))
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "log processed"})
}

func NewLogEntryHandler(svc services.LogEntryService) LogEntryHandler {
	return &logEntryHandler{svc: svc}
}
