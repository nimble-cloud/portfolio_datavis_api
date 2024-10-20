package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	ctx := context.Background()

	url := os.Getenv("DB_URL")
	fmt.Println(url)

	dbpool, err := pgxpool.New(ctx, url)
	if err != nil {
		fmt.Println("Error connecting to database")
		fmt.Println(err)
		os.Exit(1)
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		fmt.Println("Error pinging database")
		fmt.Println(err)
		os.Exit(1)
	}

	DB = dbpool
}
