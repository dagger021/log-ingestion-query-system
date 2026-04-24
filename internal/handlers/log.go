package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

	q := r.URL.Query()
	filter := services.GetLogsFilter{}

	if v := q.Get("level"); v != "" {
		filter.Level = &v
	}

	if v := q.Get("resourceId"); v != "" {
		filter.ResourceId = &v
	}

	if v := q.Get("traceId"); v != "" {
		filter.TraceId = &v
	}

	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.FromTime = &t
		}
	}

	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.ToTime = &t
		}
	}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			filter.Limit = n
		}
	}

	if v := q.Get("offset"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			filter.Offset = n
		}
	}

	logEntries, err := h.svc.GetLogs(ctx, filter)
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
