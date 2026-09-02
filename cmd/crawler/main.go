package main

import (
	"context"
	"os"
	"time"
	"web-crawler/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dburl := os.Getenv("DATABASE_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dburl)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	_ = storage.NewStorage(pool)

}
