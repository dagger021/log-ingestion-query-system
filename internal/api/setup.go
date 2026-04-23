package api

import (
	"github.com/dagger021/log-ingestion-query-system/internal/handlers"
	"github.com/go-chi/chi"
)

func SetupAPI(hdlr handlers.LogEntryHandler) chi.Router {
	router := chi.NewRouter()

	router.Get("/logs", hdlr.GetLogs)
	router.Post("/ingest", hdlr.PostLog)

	return router
}
