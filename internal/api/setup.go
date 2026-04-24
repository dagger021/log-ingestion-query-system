package api

import (
	"net/http"

	"github.com/dagger021/log-ingestion-query-system/internal/handlers"
	"github.com/dagger021/log-ingestion-query-system/internal/middleware"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func SetupAPI(logger *zap.Logger, hdlr handlers.LogEntryHandler) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Logger(logger))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	router.Get("/logs", hdlr.GetLogs)
	router.Post("/ingest", hdlr.PostLog)

	return router
}

func SetupTemplates(logger *zap.Logger, hdlr handlers.LogEntryHandler) chi.Router {
	router := chi.NewRouter()
	router.Get("/logs", hdlr.GetLogsTemplate)

	return router
}
