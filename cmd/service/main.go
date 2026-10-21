// Command service runs the webarchive service. It is the composition root: every
// implementation is built here and passed inward to the layers that declare the
// ports they use.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ogen-go/ogen/middleware"
	"go.uber.org/zap"

	"github.com/derfenix/webarchive/internal/api/openapi"
	"github.com/derfenix/webarchive/internal/application"
	"github.com/derfenix/webarchive/internal/config"
	"github.com/derfenix/webarchive/internal/infrastructure/logging"
	"github.com/derfenix/webarchive/internal/infrastructure/processors"
	"github.com/derfenix/webarchive/internal/infrastructure/queue"
	"github.com/derfenix/webarchive/internal/infrastructure/storage"
	badgerRepo "github.com/derfenix/webarchive/internal/infrastructure/storage/badger"
	"github.com/derfenix/webarchive/internal/interfaces"
	"github.com/derfenix/webarchive/internal/interfaces/rest"
	"github.com/derfenix/webarchive/internal/interfaces/web"
)

const (
	readTimeout       = 15 * time.Second
	readHeaderTimeout = 5 * time.Second
	idleTimeout       = 30 * time.Second
	shutdownTimeout   = 5 * time.Second
	maxHeaderBytes    = 1024 * 2
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "webarchive: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	log, err := logging.New(cfg.Logging)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	defer func() {
		_ = log.Sync()
	}()

	db, err := storage.Open(cfg.DB.Path, log.Named("db"))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	defer func() {
		if err := stopDB(db); err != nil {
			log.Error("failed to stop database", zap.Error(err))
		}
	}()

	pageRepo, err := badgerRepo.NewPage(db)
	if err != nil {
		return fmt.Errorf("new page repo: %w", err)
	}

	archiver, err := processors.NewProcessors(cfg.PDF, log.Named("processor"))
	if err != nil {
		return fmt.Errorf("new processors: %w", err)
	}

	// Unbuffered on purpose: adding a page waits until the worker takes the task.
	// The queue is not closed on the way out: the worker stops on the context,
	// and closing the channel under a request still pushing would panic.
	tasks := queue.NewMemory(0)

	service := application.NewService(pageRepo, archiver, tasks, log.Named("service"))
	worker := application.NewWorker(service, pageRepo, tasks, log.Named("worker"))

	server, err := openapi.NewServer(
		rest.NewHandler(service),
		openapi.WithPathPrefix(interfaces.APIPrefix),
		openapi.WithMiddleware(requestLogging(log)),
	)
	if err != nil {
		return fmt.Errorf("new rest server: %w", err)
	}

	var httpHandler http.Handler = server

	if cfg.UI.Enabled {
		httpHandler = interfaces.NewRouter(server, web.NewUI(cfg.UI))
	}

	httpServer := &http.Server{
		Addr:              cfg.API.Address,
		Handler:           httpHandler,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	return serve(ctx, log, httpServer, worker)
}

// serve runs the worker and the HTTP server and returns when both are stopped.
func serve(ctx context.Context, log *zap.Logger, server *http.Server, worker *application.Worker) error {
	runCtx, stop := context.WithCancel(ctx)
	defer stop()

	server.BaseContext = func(net.Listener) context.Context {
		return runCtx
	}

	var (
		wg       sync.WaitGroup
		serveErr error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		defer stop()

		if err := worker.Run(runCtx); err != nil {
			log.Error("worker stopped with error", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		defer stop()

		log.Info("starting http server", zap.String("address", server.Addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http serve error", zap.Error(err))

			serveErr = err
		}

		log.Info("http server stopped")
	}()

	<-runCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Warn("http graceful shutdown failed", zap.Error(err))
	}

	// Waiting for the worker keeps in-flight pages from losing their results
	// when the database is closed right after.
	wg.Wait()

	return serveErr
}

func stopDB(db *badger.DB) error {
	var errs error

	if err := db.Sync(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("sync db: %w", err))
	}

	if err := storage.Backup(db, storage.BackupStop); err != nil {
		errs = errors.Join(errs, fmt.Errorf("backup on stop: %w", err))
	}

	if err := db.Close(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("close db: %w", err))
	}

	return errs
}

func requestLogging(log *zap.Logger) openapi.Middleware {
	return func(r middleware.Request, next middleware.Next) (middleware.Response, error) {
		start := time.Now()

		log := log.With(
			zap.String("operation_id", r.OperationID),
			zap.String("uri", r.Raw.RequestURI),
		)

		response, err := next(r)

		log.Debug("request completed", zap.Duration("duration", time.Since(start)), zap.Error(err))

		return response, err
	}
}
