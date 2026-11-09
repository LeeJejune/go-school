// Package config loads and validates typed environment configuration once at startup.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	CatalogAddr    string        `env:"CATALOG_ADDR" envDefault:"127.0.0.1:9001"`
	ReadingAddr    string        `env:"READING_ADDR" envDefault:"127.0.0.1:9002"`
	HTTPAddr       string        `env:"HTTP_ADDR" envDefault:"127.0.0.1:8080"`
	CatalogTarget  string        `env:"CATALOG_TARGET" envDefault:"127.0.0.1:9001"`
	ReadingTarget  string        `env:"READING_TARGET" envDefault:"127.0.0.1:9002"`
	MySQLDSN       string        `env:"MYSQL_DSN" envDefault:"school:school@tcp(127.0.0.1:3307)/school?parseTime=true&loc=UTC&timeout=3s&readTimeout=3s&writeTimeout=3s"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" envDefault:"2s"`
	CatalogTimeout time.Duration `env:"CATALOG_TIMEOUT" envDefault:"300ms"`
	DBMaxOpen      int           `env:"DB_MAX_OPEN" envDefault:"10"`
	DBMaxIdle      int           `env:"DB_MAX_IDLE" envDefault:"5"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("load configuration: %w", err)
	}
	if cfg.RequestTimeout <= 0 || cfg.CatalogTimeout <= 0 || cfg.DBMaxOpen < 1 || cfg.DBMaxIdle < 0 || cfg.DBMaxIdle > cfg.DBMaxOpen {
		return Config{}, fmt.Errorf("timeouts must be positive and 0 <= DB_MAX_IDLE <= DB_MAX_OPEN with DB_MAX_OPEN > 0")
	}
	return cfg, nil
}
