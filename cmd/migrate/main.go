package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required (set it or put it in .env)")
	}

	sqlBytes, err := os.ReadFile("db/migrations/001_init.sql")
	if err != nil {
		log.Fatalf("failed reading migration file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("Migration applied: db/migrations/001_init.sql")
}

