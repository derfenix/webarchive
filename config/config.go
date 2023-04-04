package config

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
	API     API     `env:",prefix=API_"`
	UI      UI      `env:",prefix=UI_"`
	PDF     PDF     `env:",prefix=PDF_"`
}

type PDF struct {
	Landscape  bool    `env:"LANDSCAPE,default=false"`
	Grayscale  bool    `env:"GRAYSCALE,default=false"`
	MediaPrint bool    `env:"MEDIA_PRINT,default=true"`
	Zoom       float64 `env:"ZOOM,default=1"`
	Viewport   string  `env:"VIEWPORT,default=1280x720"`
	DPI        uint    `env:"DPI,default=150"`
	Filename   string  `env:"FILENAME,default=page.pdf"`
}

type API struct {
	Address string `env:"ADDRESS,default=0.0.0.0:5001"`
}

type UI struct {
	Enabled bool   `env:"ENABLED,default=true"`
	Prefix  string `env:"PREFIX,default=/"`
	Theme   string `env:"THEME,default=basic"`
}

type DB struct {
	Path string `env:"PATH,default=./db"`
}

type Logging struct {
	Debug bool `env:"DEBUG"`
}
