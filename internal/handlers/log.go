package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/internal/logger"
	"github.com/dagger021/log-ingestion-query-system/internal/services"
	"go.uber.org/zap"
)

type LogEntryHandler interface {
	GetLogs(http.ResponseWriter, *http.Request)

	GetLogsTemplate(http.ResponseWriter, *http.Request)

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
	filter := services.GetLogsFilter{
		Levels:      parseArray(q, "level"),
		ResourceIds: parseArray(q, "resourceId"),
		TraceIds:    parseArray(q, "traceId"),

		MessageRegex:  parseOne(q, "message"),
		MessageTokens: parseArray(q, "token"),

		FromTime: parseTime(q, "from"),
		ToTime:   parseTime(q, "to"),

		Limit:  parseUint(q, "limit"),
		Offset: parseUint(q, "offset"),
	}

	logEntries, err := h.svc.GetLogs(ctx, filter)
	if err != nil {
		log.Error("error selecting logEntries", zap.Error(err))
		WriteError(w, http.StatusInternalServerError, "invalid response payload")
		return
	}

	WriteJSON(w, http.StatusOK, logEntries)
}

func (h *logEntryHandler) GetLogsTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	q := r.URL.Query()
	filter := services.GetLogsFilter{
		Levels:      parseArray(q, "level"),
		ResourceIds: parseArray(q, "resourceId"),
		TraceIds:    parseArray(q, "traceId"),

		MessageRegex:  parseOne(q, "message"),
		MessageTokens: parseArray(q, "token"),

		FromTime: parseTime(q, "from"),
		ToTime:   parseTime(q, "to"),

		Limit:  parseUint(q, "limit"),
		Offset: parseUint(q, "offset"),
	}

	logEntries, err := h.svc.GetLogs(ctx, filter)
	if err != nil {
		log.Error("error selecting logEntries", zap.Error(err))
		WriteError(w, http.StatusInternalServerError, "invalid response payload")
		return
	}

	data := struct{ Logs any }{Logs: logEntries}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	templatePath := "templates/index.gotmpl"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Error(
			"UI template not found",
			zap.Error(err),
			zap.Any("template-path", templatePath),
		)
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
		log.Error("error rendering template", zap.Error(err), zap.Any("template-path", templatePath))
		WriteError(w, http.StatusInternalServerError, "template error: "+err.Error())
	}
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
