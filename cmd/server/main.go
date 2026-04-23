package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dagger021/log-ingestion-query-system/internal/api"
	"github.com/dagger021/log-ingestion-query-system/internal/clickhouse"
	"github.com/dagger021/log-ingestion-query-system/internal/config"
	"github.com/dagger021/log-ingestion-query-system/internal/handlers"
	"github.com/dagger021/log-ingestion-query-system/internal/kafka"
	"github.com/dagger021/log-ingestion-query-system/internal/logger"
	"github.com/dagger021/log-ingestion-query-system/internal/services"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {
	fmt.Printf("Log Ingestion & Query System\n\n")

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed initializing config: %v", err)
	}

	mainLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	chConn, err := clickhouse.InitConn(cfg.ClickhouseDSN)
	if err != nil {
		mainLogger.Fatal("failed connecting to clickhouse", zap.Error(err))
	}
	defer chConn.Close()

	// migrate file
	if err := clickhouse.Migrate(chConn, "./schema.sql"); err != nil {
		mainLogger.Fatal("failed migration", zap.Error(err))
	}

	apiLogger := mainLogger.Named("api")
	kafkaLogger := mainLogger.Named("kafka")

	kafkaBrokers := []string{cfg.KafkaBroker}
	producer := kafka.NewProducer(kafkaBrokers, kafkaLogger.Named("producer"))
	defer producer.Close()

	consumer := kafka.NewConsumer(
		kafkaBrokers, chConn, producer, kafkaLogger.Named("consumer"),
	)

	// shutdown context
	ctx, cancel := context.WithCancel(context.Background())

	// Inject services -> handlers
	logEntryService := services.NewLogEntryService(chConn)
	logEntryHandler := handlers.NewLogEntryHandler(logEntryService)

	app := chi.NewRouter()
	app.Mount("/api", api.SetupAPI(logEntryHandler))

	server := &http.Server{
		Addr:         ":3000",
		Handler:      app,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	var wg sync.WaitGroup

	wg.Go(func() { consumer.Run(ctx) })

	wg.Go(func() {
		apiLogger.Info(("server initiated @ " + server.Addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			apiLogger.Error("server failed", zap.Error(err))
			cancel()
		}
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit // wait for quit signal
	signal.Stop(quit)

	mainLogger.Info("shutdown initiated")
	cancel()

	ctxShutdown, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()

	if err := server.Shutdown(ctxShutdown); err != nil {
		apiLogger.Error("failed to shutdown server gracefully", zap.Error(err))
	}

	// wait for all go-routines
	wg.Wait()
	mainLogger.Info("shutdown complete")
}
