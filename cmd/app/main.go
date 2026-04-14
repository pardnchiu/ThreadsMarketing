package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pardnchiu/ThreadsMarketing/internal/tui"
	utils_database "github.com/pardnchiu/go-utils/database"
	utils_keychain "github.com/pardnchiu/go-utils/filesystem/keychain"
)

const (
	serviceName   = "ThreadsMarketing"
	migrationsDir = "migrations"
)

var (
	configDir = filepath.Join(os.Getenv("HOME"), ".config", serviceName)
)

func main() {
	_ = godotenv.Load()

	utils_keychain.Init(serviceName, configDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := utils_database.NewPostgresql(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres init failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := utils_database.PostgresqlMigrate(ctx, db, migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
		os.Exit(1)
	}

	app := tui.New(db)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
