package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"web-crawler/internal/crawler"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := crawler.Run(ctx, os.Args[1:], os.Getenv("DATABASE_URL"), os.Stderr)
	stop()
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		slog.Error("crawler stopped", "error", err)
		os.Exit(1)
	}
}
