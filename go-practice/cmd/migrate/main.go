package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/migrations"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command != "up" && command != "status" {
		return fmt.Errorf("usage: migrate [up|status]")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	locker, err := lock.NewMySQLTableLocker()
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectMySQL, db, migrations.Files, goose.WithLocker(locker))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if command == "status" {
		statuses, err := provider.Status(ctx)
		if err != nil {
			return err
		}
		for _, migration := range statuses {
			fmt.Printf("%s\t%s\n", migration.State, migration.Source.Path)
		}
		return nil
	}
	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}
	slog.Info("migrations applied", "count", len(results))
	return nil
}
