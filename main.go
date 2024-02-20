package main

import (
	"context"
	"fmt"
	"os"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/gapi"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("Hello, Streamfair User Management Service!")

	config, err := util.LoadConfig(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: error while loading config: %v\n", err)
		return
	}

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: error while parsing config: %v\n", err)
		return
	}

	conn, err := pgxpool.New(context.Background(), poolConfig.ConnString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connection: unable to create connection pool: %v\n", err)
		return
	}

	store := db.NewStore(conn)
	server, err := gapi.NewServer(config, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while creating server: %v\n", err)
		return
	}

	go server.RunGrpcGatewayServer()
	server.RunGrpcServer()
}
