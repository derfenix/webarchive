package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"go.uber.org/zap"

	"github.com/derfenix/webarchive/application"
	"github.com/derfenix/webarchive/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	cfg, err := config.NewConfig(ctx)
	if err != nil {
		fmt.Printf("failed to init config: %s", err.Error())
		os.Exit(2)
	}

	app, err := application.NewApplication(cfg)
	if err != nil {
		fmt.Printf("failed to init application: %s", err.Error())
		os.Exit(2)
	}

	wg := sync.WaitGroup{}

	if err := app.Start(ctx, &wg); err != nil {
		app.Log().Fatal("failed to start application", zap.Error(err))
	}

	wg.Wait()

	if err := app.Stop(); err != nil {
		app.Log().Fatal("failed to graceful stop", zap.Error(err))
	}
}
