package application

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

const envPrefix = "WEBARCHIVE_"

func NewConfig(ctx context.Context) (Config, error) {
	cfg := Config{}

	lookuper := envconfig.MultiLookuper(
		envconfig.PrefixLookuper(envPrefix, envconfig.OsLookuper()),
		envconfig.OsLookuper(),
	)

	if err := envconfig.ProcessWith(ctx, &cfg, lookuper); err != nil {
		return Config{}, fmt.Errorf("process env: %w", err)
	}

	return cfg, nil
}

type Config struct {
	DB      DB      `env:",prefix=DB_"`
	Logging Logging `env:",prefix=LOGGING_"`
}

type DB struct {
	Path string `env:"PATH,default=./db"`
}

type Logging struct {
	Debug bool `env:"DEBUG"`
}
