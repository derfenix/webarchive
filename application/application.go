package application

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ogen-go/ogen/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/derfenix/webarchive/adapters/processors"
	"github.com/derfenix/webarchive/adapters/repository"
	badgerRepo "github.com/derfenix/webarchive/adapters/repository/badger"
	"github.com/derfenix/webarchive/api/openapi"
	"github.com/derfenix/webarchive/config"
	"github.com/derfenix/webarchive/entity"
	"github.com/derfenix/webarchive/ports/rest"
)

func NewApplication(cfg config.Config) (Application, error) {
	log, err := newLogger(cfg.Logging)
	if err != nil {
		return Application{}, fmt.Errorf("new logger: %w", err)
	}

	db, err := repository.NewBadger(cfg.DB.Path, log.Named("db"))
	if err != nil {
		return Application{}, fmt.Errorf("new badger: %w", err)
	}

	pageRepo, err := badgerRepo.NewPage(db)
	if err != nil {
		return Application{}, fmt.Errorf("new page repo: %w", err)
	}

	processor, err := processors.NewProcessors(cfg, log.Named("processor"))
	if err != nil {
		return Application{}, fmt.Errorf("new processors: %w", err)
	}

	workerCh := make(chan *entity.Page)
	worker := entity.NewWorker(workerCh, pageRepo, processor, log.Named("worker"))

	server, err := openapi.NewServer(
		rest.NewService(pageRepo, workerCh, processor),
		openapi.WithPathPrefix("/api/v1"),
		openapi.WithMiddleware(
			func(r middleware.Request, next middleware.Next) (middleware.Response, error) {
				start := time.Now()

				log := log.With(
					zap.String("operation_id", r.OperationID),
					zap.String("uri", r.Raw.RequestURI),
				)

				var response middleware.Response
				var reqErr error

				response, reqErr = next(r)

				log.Debug("request completed", zap.Duration("duration", time.Since(start)), zap.Error(err))

				return response, reqErr
			},
		),
	)
	if err != nil {
		return Application{}, fmt.Errorf("new rest server: %w", err)
	}

	var httpHandler http.Handler = server

	if cfg.UI.Enabled {
		ui := rest.NewUI(cfg.UI)

		httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				server.ServeHTTP(w, r)

				return
			}

			ui.ServeHTTP(w, r)
		})
	}

	httpServer := http.Server{
		Addr:              cfg.API.Address,
		Handler:           httpHandler,
		ReadTimeout:       time.Second * 15,
		ReadHeaderTimeout: time.Second * 5,
		IdleTimeout:       time.Second * 30,
		MaxHeaderBytes:    1024 * 2,
	}

	return Application{
		cfg:        cfg,
		log:        log,
		db:         db,
		processor:  processor,
		httpServer: &httpServer,
		worker:     worker,

		pageRepo: pageRepo,
	}, nil
}

type Application struct {
	cfg        config.Config
	log        *zap.Logger
	db         *badger.DB
	processor  entity.Processor
	httpServer *http.Server
	worker     *entity.Worker

	pageRepo *badgerRepo.Page
}

func (a *Application) Log() *zap.Logger {
	return a.log
}

func (a *Application) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(3)

	a.httpServer.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	go a.worker.Start(ctx, wg)

	go func() {
		defer wg.Done()

		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			a.log.Warn("http graceful shutdown failed", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()

		a.log.Info("starting http server", zap.String("address", a.httpServer.Addr))

		if err := a.httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.log.Error("http serve error", zap.Error(err))
			}

			a.log.Info("http server stopped")
		}
	}()

	return nil
}

func (a *Application) Stop() error {
	var errs error

	if err := a.db.Sync(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("sync db: %w", err))
	}

	if err := repository.Backup(a.db, repository.BackupStop); err != nil {
		errs = errors.Join(errs, fmt.Errorf("backup on stop: %w", err))
	}

	if err := a.db.Close(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("close db: %w", err))
	}

	return errs
}

func newLogger(cfg config.Logging) (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	logCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logCfg.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	logCfg.DisableCaller = true
	logCfg.DisableStacktrace = true

	logCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if cfg.Debug {
		logCfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	log, err := logCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return log, nil
}
