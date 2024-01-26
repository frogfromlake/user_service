package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Streamfair/streamfair_user_svc/api"
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("Hello, Streamfair User Service!")
	config, err := util.LoadConfig(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: error while loading config: %v\n", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: error while parsing config: %v\n", err)
	}

	conn, err := pgxpool.New(context.Background(), poolConfig.ConnString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connection: unable to create connection pool: %v\n", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.StartServer(config.ServerAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while starting server: %v\n", err)
	}
}
