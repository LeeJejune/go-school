package data

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// NewDB provides one shared GORM handle/connection pool for the application.
// Schema changes belong to the separate Goose migration command, not AutoMigrate.
func NewDB(lc fx.Lifecycle, cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: cfg.MySQLDSN, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger: gormlog.New(log.New(os.Stderr, "gorm ", log.LstdFlags), gormlog.Config{
			SlowThreshold: 200 * time.Millisecond, LogLevel: gormlog.Warn,
			IgnoreRecordNotFoundError: true, ParameterizedQueries: true,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("open GORM: %w", err)
	}
	pool, err := db.DB()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(cfg.DBMaxOpen)
	pool.SetMaxIdleConns(cfg.DBMaxIdle)
	pool.SetConnMaxLifetime(3 * time.Minute)
	pool.SetConnMaxIdleTime(time.Minute)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.PingContext(ctx); err != nil {
				_ = pool.Close()
				return fmt.Errorf("ping MySQL: %w", err)
			}
			return nil
		},
		OnStop: func(context.Context) error { return pool.Close() },
	})
	return db, nil
}
