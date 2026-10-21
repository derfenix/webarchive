// Package logging builds the service logger.
package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/derfenix/webarchive/internal/config"
)

func New(cfg config.Logging) (*zap.Logger, error) {
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
