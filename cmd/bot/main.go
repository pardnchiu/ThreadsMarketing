package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	scheduler "github.com/pardnchiu/go-scheduler"
	utils_database "github.com/pardnchiu/go-utils/database"

	"github.com/pardnchiu/ThreadsMarketing/internal/bot"
)

const (
	migrationsDir    = "migrations"
	collectInterval  = "@every 1m"
	downloadInterval = "@every 30s"
)

func main() {
	_ = godotenv.Load()

	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := utils_database.NewPostgresql(initCtx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres init failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := utils_database.PostgresqlMigrate(initCtx, db, migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
		os.Exit(1)
	}

	logger := func(msg string) { log.Println(msg) }
	b := bot.New(db, bot.Logger(logger))
	defer b.Close()

	location, _ := time.LoadLocation("Asia/Taipei")
	cron, err := scheduler.New(scheduler.Config{Location: location})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scheduler init: %v\n", err)
		os.Exit(1)
	}

	if _, err := cron.Add(collectInterval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		b.CollectURLs(ctx)
	}, "collector"); err != nil {
		fmt.Fprintf(os.Stderr, "add collector: %v\n", err)
		os.Exit(1)
	}

	if _, err := cron.Add(downloadInterval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		b.Download(ctx)
	}, "downloader"); err != nil {
		fmt.Fprintf(os.Stderr, "add downloader: %v\n", err)
		os.Exit(1)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		b.CollectURLs(ctx)
	}()

	cron.Start()
	log.Println("[bot] started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("[bot] shutting down")
}
