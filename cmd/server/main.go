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
	"github.com/dagger021/log-ingestion-query-system/internal/services"
	"github.com/go-chi/chi"
)

func main() {
	fmt.Printf("Log Ingestion & Query System\n\n")

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed initializing config")
	}

	chConn, err := clickhouse.InitConn(cfg.ClickhouseDSN)
	if err != nil {
		log.Fatalf("failed to connect to clickhouse: %s", err)
	}
	defer chConn.Close()

	// migrate file
	if err := clickhouse.Migrate(chConn, "./schema.sql"); err != nil {
		log.Fatalf("failed to migirate: %s", err)
	}

	kafkaBrokers := []string{cfg.KafkaBroker}
	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	consumer := kafka.NewConsumer(kafkaBrokers, chConn, producer)

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
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("server failed:", err)
			cancel()
		}
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit // wait for quit signal
	signal.Stop(quit)

	log.Printf("shutdown initiated")
	cancel()

	ctxShutdown, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Printf("failed to shutdown server gracefully: %s", err)
	}

	// wait for all go-routines
	wg.Wait()
}
