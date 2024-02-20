package main

import (
	"context"
	"fmt"
	"log"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/gapi"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("Hello, Streamfair User Management Service!")

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Printf("config: error while loading config: %v\n", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		log.Printf("config: error while parsing config: %v\n", err)
	}

	conn, err := pgxpool.New(context.Background(), poolConfig.ConnString())
	if err != nil {
		log.Printf("db connection: unable to create connection pool: %v\n", err)
	}

	store := db.NewStore(conn)
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Printf("server: error while creating server: %v\n", err)
	}

	go server.RunGrpcGatewayServer()
	server.RunGrpcServer()
}
